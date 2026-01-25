package note

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	noteCreatorApp  *notecreator.Service
	noteInteractApp *noteinteract.Service
}

func NewHandler(c *config.Config, bizz *biz.Biz, manager *app.Manager) *Handler {
	return &Handler{
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
