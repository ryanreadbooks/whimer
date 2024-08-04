package backend

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/comment/sdk"
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

// 只获取主评论
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
		var replies = []*comment.ReplyItem{}
		if len(rootReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, rootReplies.Replies)
			if err != nil {
				httpx.Error(w, err)
				return
			}
		}

		httpx.OkJson(w, &comment.CommentRes{
			Replies:    replies,
			NextCursor: rootReplies.NextCursor,
			HasNext:    rootReplies.HasNext,
		})
	}
}

// 只获取子评论
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

		// 填充评论的用户信息
		var replies = []*comment.ReplyItem{}
		if len(subReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, subReplies.Replies)
			if err != nil {
				httpx.Error(w, err)
				return
			}
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
		var req comment.GetCommentsReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := comment.GetCommenter().
			PageGetDetailedReply(ctx, req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		var replies = []*comment.DetailedReplyItem{}
		if len(resp.Replies) > 0 {
			uidsMap := make(map[uint64]struct{})
			// 提取出主评论和子评论的uid
			for _, item := range resp.Replies {
				uidsMap[item.Root.Id] = struct{}{}
				for _, sub := range item.Subreplies {
					uidsMap[sub.Uid] = struct{}{}
				}
			}

			// 发起请求获取uid的详细信息
			userResp, err := passport.GetUserer().
				BatchGetUser(ctx, &user.BatchGetUserReq{Uids: maps.Keys(uidsMap)})
			if err != nil {
				httpx.Error(w, err)
				return
			}

			// 拼接结果
			replies = make([]*comment.DetailedReplyItem, 0, len(resp.Replies))
			for _, item := range resp.Replies {
				detail := &comment.DetailedReplyItem{}
				detail.Root = &comment.ReplyItem{
					ReplyItem: item.Root,
					User:      userResp.Users[formatUid(item.Root.Uid)],
				}
				detail.SubReplies = []*comment.ReplyItem{}
				for _, sub := range item.Subreplies {
					detail.SubReplies = append(detail.SubReplies, &comment.ReplyItem{
						ReplyItem: sub,
						User:      userResp.Users[formatUid(sub.Uid)],
					})
				}
				replies = append(replies, detail)
			}
		}

		httpx.OkJson(w, &comment.DetailedCommentRes{
			Replies:    replies,
			HasNext:    resp.HasNext,
			NextCursor: resp.NextCursor,
		})
	}
}

// 填入用户信息
func attachReplyUsers(ctx context.Context, replies []*sdk.ReplyItem) ([]*comment.ReplyItem, error) {
	uidsMap := make(map[uint64]struct{})
	for _, root := range replies {
		uidsMap[root.Uid] = struct{}{}
	}

	userResp, err := passport.GetUserer().
		BatchGetUser(ctx, &user.BatchGetUserReq{Uids: maps.Keys(uidsMap)})
	if err != nil {
		return nil, err
	}

	res := make([]*comment.ReplyItem, 0, len(replies))
	for _, root := range replies {
		res = append(res, &comment.ReplyItem{
			ReplyItem: root,
			User:      userResp.Users[formatUid(root.Uid)],
		})
	}

	return res, nil
}

func formatUid(uid uint64) string {
	return strconv.FormatUint(uid, 10)
}

func (h *Handler) DelComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.DelReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		_, err := comment.GetCommenter().DelReply(r.Context(), &sdk.DelReplyReq{ReplyId: req.ReplyId})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) PinComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req comment.PinReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		_, err := comment.GetCommenter().PinReply(r.Context(), &sdk.PinReplyReq{
			Oid:    req.Oid,
			Rid:    req.ReplyId,
			Action: sdk.ReplyAction(req.Action),
		})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}
