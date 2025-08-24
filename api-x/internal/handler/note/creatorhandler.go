package note

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) creatorSyncNoteToSearcher(ctx context.Context, noteId int64) {
	// add to es
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("creator_sync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			// 1. fetch first
			curNote, err := infra.NoteCreatorServer().GetNote(ctx,
				&notev1.GetNoteRequest{
					NoteId: noteId,
				})
			if err != nil {
				xlog.Msg("creator sync note to searcher failed").Err(err).Extras("note_id", noteId).Errorx(ctx)
				return xerror.Wrapf(err, "get note failed").WithExtra("note_id", noteId).WithCtx(ctx)
			}

			if curNote.Note.GetPrivacy() == VisibilityPrivate {
				return nil
			}

			// 2. add to searcher
			nid := model.NoteId(noteId).String()
			tagList := make([]*searchv1.NoteTag, 0, len(curNote.Note.GetTags()))
			for _, tag := range curNote.Note.GetTags() {
				tagId := model.TagId(tag.GetId()).String()
				tagList = append(tagList, &searchv1.NoteTag{
					Id:    string(tagId),
					Name:  tag.GetName(),
					Ctime: tag.GetCtime(),
				})
			}

			vis := searchv1.Note_VISIBILITY_PUBLIC
			if curNote.Note.GetPrivacy() == VisibilityPrivate {
				vis = searchv1.Note_VISIBILITY_PRIVATE
			}
			assetType := searchv1.Note_ASSET_TYPE_IMAGE // for now

			docNote := []*searchv1.Note{{
				NoteId:   string(nid),
				Title:    curNote.Note.GetTitle(),
				Desc:     curNote.Note.GetDesc(),
				CreateAt: curNote.Note.GetCreateAt(),
				UpdateAt: curNote.Note.GetUpdateAt(),
				Author: &searchv1.Note_Author{
					Uid:      curNote.Note.GetOwner(),
					Nickname: metadata.UserNickname(ctx),
				},
				TagList:    tagList,
				Visibility: vis,
				AssetType:  assetType,
			}}

			_, err = infra.DocumentServer().BatchAddNote(ctx, &searchv1.BatchAddNoteRequest{
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

func (h *Handler) creatorUnSyncNoteToSearcher(ctx context.Context, noteId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: fmt.Sprintf("creator_unsync_note_%d", noteId),
		Job: func(ctx context.Context) error {
			// 1. fetch first
			_, err := infra.DocumentServer().BatchDeleteNote(ctx, &searchv1.BatchDeleteNoteRequest{
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

// 发布笔记
func (h *Handler) CreatorCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		// service to create note
		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().CreateNote(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.creatorSyncNoteToSearcher(ctx, resp.NoteId)

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
		_, err = infra.NoteCreatorServer().UpdateNote(ctx, &notev1.UpdateNoteRequest{
			NoteId: int64(req.NoteId),
			Note: &notev1.CreateNoteRequest{
				Basic:  req.Basic.AsPb(),
				Images: req.Images.AsPb(),
			},
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.creatorSyncNoteToSearcher(ctx, int64(req.NoteId))

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
		_, err = infra.NoteCreatorServer().DeleteNote(ctx, &notev1.DeleteNoteRequest{
			NoteId: int64(req.NoteId),
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.creatorUnSyncNoteToSearcher(ctx, int64(req.NoteId))

		httpx.OkJson(w, nil)
	}
}

// 列出个人笔记
func (h *Handler) CreatorListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().ListNote(ctx, &notev1.ListNoteRequest{
			Cursor: req.Cursor,
			Count:  req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		result := NewListResFromPb(resp)
		h.assignNoteExtra(ctx, result.Items)

		xhttp.OkJson(w, result)
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
		resp, err := infra.NoteCreatorServer().PageListNote(ctx, &notev1.PageListNoteRequest{
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
		resp, err := infra.NoteCreatorServer().GetNote(ctx, &notev1.GetNoteRequest{
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
// Use [CreatorUploadNoteAuthV2] instead
func (h *Handler) CreatorUploadNoteAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := infra.NoteCreatorServer().BatchGetUploadAuth(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) CreatorUploadNoteAuthV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := infra.NoteCreatorServer().BatchGetUploadAuthV2(ctx, req.AsPbV2())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
