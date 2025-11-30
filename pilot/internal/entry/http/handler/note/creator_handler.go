package note

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/storage"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) fetchNote(ctx context.Context, noteId int64) (*notev1.NoteItem, error) {
	curNote, err := dep.NoteCreatorServer().GetNote(ctx,
		&notev1.GetNoteRequest{
			NoteId: noteId,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "get note failed").WithExtra("note_id", noteId).WithCtx(ctx)
	}

	return curNote.GetNote(), nil
}

func isNotePrivate(note *notev1.NoteItem) bool {
	return note.GetPrivacy() == int32(notev1.NotePrivacy_PRIVATE)
}

func (h *Handler) creatorSyncNoteToSearcher(ctx context.Context, noteId int64, note *notev1.NoteItem) {
	// add to es
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("note_creator_sync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			if isNotePrivate(note) {
				return nil
			}

			// 2. add to searcher
			nid := model.NoteId(noteId).String()
			tagList := make([]*searchv1.NoteTag, 0, len(note.GetTags()))
			for _, tag := range note.GetTags() {
				tagId := model.TagId(tag.GetId()).String()
				tagList = append(tagList, &searchv1.NoteTag{
					Id:    string(tagId),
					Name:  tag.GetName(),
					Ctime: tag.GetCtime(),
				})
			}

			vis := searchv1.Note_VISIBILITY_PUBLIC
			if isNotePrivate(note) {
				vis = searchv1.Note_VISIBILITY_PRIVATE
			}
			assetType := searchv1.Note_ASSET_TYPE_IMAGE // for now

			docNote := []*searchv1.Note{{
				NoteId:   string(nid),
				Title:    note.GetTitle(),
				Desc:     note.GetDesc(),
				CreateAt: note.GetCreateAt(),
				UpdateAt: note.GetUpdateAt(),
				Author: &searchv1.Note_Author{
					Uid:      note.GetOwner(),
					Nickname: metadata.UserNickname(ctx),
				},
				TagList:       tagList,
				Visibility:    vis,
				AssetType:     assetType,
				LikesCount:    note.Likes,
				CommentsCount: note.Replies,
			}}

			_, err := dep.DocumentServer().BatchAddNote(ctx, &searchv1.BatchAddNoteRequest{
				Notes: docNote,
			})
			if err != nil {
				xlog.Msg("creator sync note to searcher failed").Err(err).Extras("note_id", noteId).Errorx(ctx)
				return xerror.Wrapf(err, "batch add note failed").WithExtra("note_id", noteId).WithCtx(ctx)
			}

			return nil
		},
	})
}

func (h *Handler) creatorDeleteNoteFromSearcher(ctx context.Context, noteId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("creator_unsync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			_, err := dep.DocumentServer().BatchDeleteNote(ctx, &searchv1.BatchDeleteNoteRequest{
				Ids: []string{model.NoteId(noteId).String()},
			})
			if err != nil {
				xlog.Msg("creator unsync note to searcher failed").Err(err).Extras("note_id", noteId).Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

func (h *Handler) afterNoteUpserted(ctx context.Context, note *notev1.NoteItem) {
	h.creatorSyncNoteToSearcher(ctx, note.NoteId, note)
	h.notifyWhenAtUsers(ctx, note)
	h.appendRecentContacts(ctx, note)
}

func (h *Handler) furthurCheckNoteImages(ctx context.Context, req *CreateReq) error {
	var (
		keys        = make([]string, 0, len(req.Images))
		bucket, key string
		err         error
	)

	for _, img := range req.Images {
		bucket, key, err = h.storageBiz.SeperateResource(uploadresource.NoteImage, img.FileId)
		if err != nil {
			return err
		}

		keys = append(keys, key)
	}

	// 确保上传的资源此时是存在的
	_, err = h.storageBiz.BatchCheckResourceExist(ctx, bucket, keys, true)
	if err != nil {
		if errors.Is(err, bizstorage.ErrResourceNotFound) {
			err = modelerr.ErrResourceNotFound
		}
	} else {
		// 进一步检查所有资源都是合法的受支持的图片格式
		shouldTag := true
		for _, key := range keys {
			var (
				content []byte
				total   int64
			)
			content, total, err = h.storageBiz.GetResourceBytes(ctx, bucket, key, 32) // in bytes
			if err != nil {
				if errors.Is(err, bizstorage.ErrResourceNotFound) {
					err = modelerr.ErrResourceNotFound
				}
				shouldTag = false
				break
			}

			if err = uploadresource.NoteImage.Check(content, total); err != nil {
				shouldTag = false
				break
			}
		}

		if shouldTag {
			// 尽量给object打标签 方便后续清理脏数据
			// 后续扫描bucket时发现不存在tag的就可以考虑删除
			h.storageBiz.BatchMarkResourceActive(ctx, bucket, keys, false) // ignore error here
		}
	}

	return err
}

func (h *Handler) furthurCheckNoteVideo(ctx context.Context, req *CreateReq) error {
	// TODO
	return nil
}

// 发布笔记
func (h *Handler) CreatorCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}
		ctx := r.Context()

		// check every resource
		switch req.Basic.Type {
		case model.NoteTypeImage:
			err = h.furthurCheckNoteImages(ctx, req)
		case model.NoteTypeVideo:
			err = h.furthurCheckNoteVideo(ctx, req)
		}

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// service to create note
		resp, err := dep.NoteCreatorServer().CreateNote(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		note, err := h.fetchNote(ctx, resp.NoteId)
		if err == nil && note != nil {
			h.afterNoteUpserted(ctx, note)
		}

		xhttp.OkJson(w, CreateRes{NoteId: model.NoteId(resp.NoteId)})
	}
}

// 修改笔记
func (h *Handler) CreatorUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		_, err = dep.NoteCreatorServer().UpdateNote(ctx, &notev1.UpdateNoteRequest{
			NoteId: int64(req.NoteId),
			Note:   req.CreateReq.AsPb(),
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		note, err := h.fetchNote(ctx, int64(req.NoteId))
		if err == nil && note != nil {
			h.afterNoteUpserted(ctx, note)
		}

		xhttp.OkJson(w, UpdateRes{NoteId: req.NoteId})
	}
}

// 删除笔记
func (h *Handler) CreatorDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		_, err = dep.NoteCreatorServer().DeleteNote(ctx, &notev1.DeleteNoteRequest{
			NoteId: int64(req.NoteId),
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.creatorDeleteNoteFromSearcher(ctx, int64(req.NoteId))
		// 这里不用主动删除系统消息 在获取系统消息时延迟删除

		httpx.OkJson(w, nil)
	}
}

// 分页列出个人笔记
func (h *Handler) CreatorPageListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PageListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := dep.NoteCreatorServer().PageListNote(ctx, &notev1.PageListNoteRequest{
			Page:  req.Page,
			Count: req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := NewPageListResFromPb(resp)
		h.assignNoteExtra(ctx, result.Items)

		xhttp.OkJson(w, result)
	}
}

// 获取个人笔记
func (h *Handler) CreatorGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := dep.NoteCreatorServer().GetNote(ctx, &notev1.GetNoteRequest{
			NoteId: int64(req.NoteId),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := model.NewAdminNoteItemFromPb(resp.Note)
		h.assignNoteExtra(ctx, []*model.AdminNoteItem{result})
		xhttp.OkJson(w, result)
	}
}

// Deprecated
//
// See: upload.GetTempCreds for newest usage
func (h *Handler) CreatorUploadNoteAuthV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := h.storageBiz.RequestUploadTicket(ctx, uploadresource.NoteImage, req.Count, req.Source)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 笔记中存在at_users通知用户
func (h *Handler) notifyWhenAtUsers(ctx context.Context, note *notev1.NoteItem) {
	if isNotePrivate(note) {
		return
	}

	if len(note.AtUsers) == 0 {
		return
	}
	var (
		uid = metadata.Uid(ctx)
	)

	atUids := make([]int64, 0, len(note.AtUsers))
	for _, atUser := range note.AtUsers {
		atUids = append(atUids, atUser.Uid)
	}
	atUids = xslice.Uniq(atUids)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "note_creator_notify_at_users",
		Job: func(ctx context.Context) error {
			// filter invalid uids
			atUsers, err := h.userBiz.ListUsersV2(ctx, atUids)
			if err != nil {
				xlog.Msg("note creator user biz list users failed").Err(err).Errorx(ctx)
				return err
			}

			if len(atUsers) == 0 {
				xlog.Msg("user biz return empty at users").Errorx(ctx)
				return nil
			}

			validNoteAtUsers := make([]*notev1.NoteAtUser, 0, len(note.AtUsers))
			for _, noteAtUser := range note.AtUsers {
				if _, ok := atUsers[noteAtUser.Uid]; ok {
					validNoteAtUsers = append(validNoteAtUsers, noteAtUser)
				}
			}

			err = h.notifyBiz.NotifyAtUsersOnNote(ctx, &bizsysnotify.NotifyAtUsersOnNoteReq{
				Uid:         uid,
				TargetUsers: validNoteAtUsers,
				Content: &bizsysnotify.NotifyAtUsersOnNoteReqContent{
					NoteDesc:  note.Desc,
					NoteId:    model.NoteId(note.NoteId),
					SourceUid: uid,
				},
			})
			if err != nil {
				xlog.Msg("note creator notify biz failed to notify at users").Err(err).Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

// 异步写入最近联系人
func (h *Handler) appendRecentContacts(ctx context.Context, note *notev1.NoteItem) {
	var (
		uid = metadata.Uid(ctx)
	)

	atUsers := model.AtUserList{}
	for _, atUser := range note.GetAtUsers() {
		if atUser.Uid == metadata.Uid(ctx) {
			continue
		}

		atUsers = append(atUsers, model.AtUser{
			Uid:      atUser.Uid,
			Nickname: atUser.Nickname,
		})
	}

	h.userBiz.AsyncAppendRecentContactsAtUser(ctx, uid, atUsers)
}
