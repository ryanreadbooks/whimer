package note

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	biznote "github.com/ryanreadbooks/whimer/pilot/internal/biz/note"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	noteCreatorApp  *notecreator.Service
	noteInteractApp *noteinteract.Service

	feedBiz    *bizfeed.Biz
	searchBiz  *bizsearch.Biz
	notifyBiz  *bizsysnotify.Biz
	storageBiz *bizstorage.Biz
	noteBiz    *biznote.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz, manager *app.Manager) *Handler {
	return &Handler{
		feedBiz:         bizz.FeedBiz,
		searchBiz:       bizz.SearchBiz,
		notifyBiz:       bizz.SysNotifyBiz,
		storageBiz:      bizz.UploadBiz,
		noteBiz:         bizz.NoteBiz,
		noteCreatorApp:  manager.NoteCreatorApp,
		noteInteractApp: manager.NoteInteractApp,
	}
}

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.LikeNoteCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		err = h.noteInteractApp.LikeNote(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// TODO 通知用户笔记被点赞了
		// if req.Action == imodel.LikeReqActionDo {
		// 	h.asyncNotifyLikeNote(ctx, noteId)
		// }

		xhttp.OkJson(w, nil)
	}
}

// 获取笔记点赞数量
func (h *Handler) GetNoteLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commondto.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		cnt, err := h.noteInteractApp.GetLikeCount(r.Context(), req.NoteId.Int64())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &dto.GetLikeCountResult{
			NoteId: req.NoteId,
			Count:  cnt,
		})
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
