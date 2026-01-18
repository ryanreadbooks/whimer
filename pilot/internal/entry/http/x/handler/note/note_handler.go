package note

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	feedmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed/model"
	biznote "github.com/ryanreadbooks/whimer/pilot/internal/biz/note"
	notemodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/note/model"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	feedBiz    *bizfeed.Biz
	searchBiz  *bizsearch.Biz
	userBiz    *bizuser.Biz
	notifyBiz  *bizsysnotify.Biz
	storageBiz *bizstorage.Biz
	noteBiz    *biznote.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		feedBiz:    bizz.FeedBiz,
		searchBiz:  bizz.SearchBiz,
		userBiz:    bizz.UserBiz,
		notifyBiz:  bizz.SysNotifyBiz,
		storageBiz: bizz.UploadBiz,
		noteBiz:    bizz.NoteBiz,
	}
}

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		noteId := req.NoteId
		noteIdStr := noteId.String()

		err = h.noteBiz.LikeNote(ctx, &notemodel.LikeNoteReq{
			NoteId: noteId,
			Action: req.Action,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "note.handler.likenote.synces",
			Job: func(ctx context.Context) error {
				var incr int64 = 1
				if req.Action == imodel.LikeReqActionUndo {
					incr = -1
				}

				err := h.searchBiz.NoteStatSyncer.AddLikeCount(ctx, noteIdStr, incr)
				if err != nil {
					xlog.Msg("note stat add like count failed").
						Extras("note_id", noteId, "note_id_str", noteIdStr).
						Err(err).Errorx(ctx)
				}

				return err
			},
		})

		if req.Action == imodel.LikeReqActionDo {
			h.asyncNotifyLikeNote(ctx, noteId)
		}

		xhttp.OkJson(w, nil)
	}
}

// 获取笔记点赞数量
func (h *Handler) GetNoteLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := h.noteBiz.GetNoteLikeCount(r.Context(), req.NoteId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetLikesRes{
			NoteId: resp.NoteId,
			Count:  resp.Count,
		})
	}
}

// 创建新标签
func (h *Handler) AddNewTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[AddTagReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.noteBiz.AddTag(ctx, req.Name)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.noteBiz.AsyncTagToSearcher(ctx, int64(resp.TagId))

		xhttp.OkJson(w, &AddTagRes{TagId: resp.TagId})
	}
}

// 搜索笔记标签
func (h *Handler) SearchTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SearchTagsReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if req.Name == "" {
			xhttp.OkJson(w, []SearchTagsRes{})
			return
		}

		tags, err := h.noteBiz.SearchTags(r.Context(), req.Name)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := make([]SearchedNoteTag, len(tags))
		for idx, tag := range tags {
			result[idx].Id = tag.Id
			result[idx].Name = tag.Name
		}

		xhttp.OkJson(w, result)
	}
}

// 获取用户点赞过的笔记
func (b *Handler) ListLikedNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetLikedNoteRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx    = r.Context()
			curUid = metadata.Uid(ctx)
		)

		if curUid != req.Uid {
			// 检查req.Uid的点赞记录是否公开
			userSetting, _ := b.userBiz.GetIntegralUserSettings(ctx, req.Uid)
			if !userSetting.ShowNoteLikes {
				xhttp.Error(r, w, modelerr.ErrLikesHistoryHidden)
				return
			}
		}

		noteResp, err := dep.NoteInteractServer().PageListUserLikedNote(ctx,
			&notev1.PageListUserLikedNoteRequest{
				Uid:    req.Uid,
				Cursor: req.Cursor,
				Count:  req.Count,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp := &GetLikedNoteResponse{
			Items:      []*feedmodel.FeedNoteItem{},
			NextCursor: noteResp.NextCursor,
			HasNext:    noteResp.HasNext,
		}

		if len(noteResp.Items) > 0 {
			targets, err := b.feedBiz.AssembleNoteFeeds(ctx, noteResp.GetItems())
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
			resp.Items = targets
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) asyncNotifyLikeNote(ctx context.Context, noteId imodel.NoteId) {
	uid := metadata.Uid(ctx)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "note.handler.notify_like_note",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			author, err := h.feedBiz.GetNoteAuthor(ctx, int64(noteId))
			if err != nil {
				return xerror.Wrapf(err, "feed biz get note author failed").WithExtra("note_id", noteId).WithCtx(ctx)
			}

			err = h.notifyBiz.NotifyUserLikesOnNote(ctx, uid, author, &bizsysnotify.NotifyLikesOnNoteReq{
				NoteId: noteId,
			})
			if err != nil {
				return xerror.Wrapf(err, "notify likes on note failed").
					WithExtras("note_id", noteId, "uid", uid, "recv", author).WithCtx(ctx)
			}

			return nil
		},
	})
}
