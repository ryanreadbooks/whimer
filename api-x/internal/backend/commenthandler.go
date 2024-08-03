package backend

import (
	"net/http"
	"strconv"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/passport/sdk/user"

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

func (h *Handler) PageGetRoots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.GetCommentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		rootReplies, err := comment.GetCommenter().
			PageGetReply(ctx, req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// 整合用户的信息
		uidsMap := make(map[uint64]struct{})
		for _, root := range rootReplies.Replies {
			uidsMap[root.Uid] = struct{}{}
		}

		// 获取用户信息
		userResp, err := passport.GetUserer().
			BatchGetUser(ctx, &user.BatchGetUserReq{Uids: maps.Keys(uidsMap)})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		replies := make([]*comment.ReplyItem, 0, len(rootReplies.Replies))
		for _, root := range rootReplies.Replies {
			replies = append(replies, &comment.ReplyItem{
				ReplyItem: root,
				User:      userResp.Users[strconv.FormatUint(root.Uid, 10)],
			})
		}

		httpx.OkJson(w, &comment.CommentRes{
			Replies:    replies,
			NextCursor: rootReplies.NextCursor,
			HasNext:    rootReplies.HasNext,
		})
	}
}

func (h *Handler) PageGetSubs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.GetSubCommentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		subReplies, err := comment.GetCommenter().
			PageGetSubReply(ctx, req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// 整合用户的信息
		uidsMap := make(map[uint64]struct{})
		for _, root := range subReplies.Replies {
			uidsMap[root.Uid] = struct{}{}
		}

		userResp, err := passport.GetUserer().
			BatchGetUser(ctx, &user.BatchGetUserReq{Uids: maps.Keys(uidsMap)})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		replies := make([]*comment.ReplyItem, 0, len(subReplies.Replies))
		for _, root := range subReplies.Replies {
			replies = append(replies, &comment.ReplyItem{
				ReplyItem: root,
				User:      userResp.Users[strconv.FormatUint(root.Uid, 10)],
			})
		}

		httpx.OkJson(w, &comment.CommentRes{
			Replies:    replies,
			HasNext:    subReplies.HasNext,
			NextCursor: subReplies.NextCursor,
		})
	}
}

// 获取主评论信息（包含其下子评论）
func (h *Handler) PageGetReplies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
