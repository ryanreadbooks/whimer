package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/misc/errorx"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 发表评论
func (h *Handler) PublishComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.PubReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := comment.GetCommenter().
			AddReply(r.Context(), req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, &comment.PubRes{ReplyId: resp.ReplyId})
	}
}

func (h *Handler) PageGetComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.GetCommentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		rootReplies, err := comment.GetCommenter().
			PageGetReply(r.Context(), req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}
		
		// TODO 获取主评论下的子评论

		// TODO 整合每个用户的信息

		httpx.OkJson(w, rootReplies)
	}
}
