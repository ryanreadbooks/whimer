package backend

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 发表评论
func (h *Handler) PublishComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.PubReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := comment.Commenter().
			AddReply(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &comment.PubRes{ReplyId: resp.ReplyId})
	}
}

// 只获取主评论
func (h *Handler) PageGetRoots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		rootReplies, err := comment.Commenter().
			PageGetReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 整合用户的信息
		var replies = []*comment.ReplyItem{}
		if len(rootReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, rootReplies.Replies)
			if err != nil {
				xhttp.Error(r, w, err)
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
		req, err := xhttp.ParseValidate[comment.GetSubCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		subReplies, err := comment.Commenter().
			PageGetSubReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 填充评论的用户信息
		var replies = []*comment.ReplyItem{}
		if len(subReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, subReplies.Replies)
			if err != nil {
				xhttp.Error(r, w, err)
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

func extractUidsMap(replies []*commentv1.DetailedReplyItem) map[uint64]struct{} {
	uidsMap := make(map[uint64]struct{})
	// 提取出主评论和子评论的uid
	for _, item := range replies {
		uidsMap[item.Root.Uid] = struct{}{}
		for _, sub := range item.Subreplies.Items {
			uidsMap[sub.Uid] = struct{}{}
		}
	}

	return uidsMap
}

// 获取主评论信息（包含其下子评论）
func (h *Handler) PageGetReplies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			pinnedReply     *commentv1.DetailedReplyItem
			pinnedReplyUser map[string]*userv1.UserInfo = nil
			wg              sync.WaitGroup
			ctx             = r.Context()
		)

		if req.Cursor == 0 {
			wg.Add(1)
			// 第一次请求时需要返回置顶评论
			concurrent.SafeGo(func() {
				defer wg.Done()
				var err error
				resp, err := comment.Commenter().
					GetPinnedReply(ctx, &commentv1.GetPinnedReplyReq{Oid: req.Oid})
				if err != nil {
					logx.Errorw("rpc get pin reply err", xlog.WithUid(ctx), xlog.WithErr(err))
					return
				}
				pinnedReply = resp.Reply

				userResp, err := passport.Userer().
					BatchGetUser(ctx,
						&userv1.BatchGetUserRequest{
							Uids: maps.Keys(extractUidsMap([]*commentv1.DetailedReplyItem{pinnedReply})),
						},
					)
				if err != nil {
					logx.Errorw("rpc get batch get user err", xlog.WithUid(ctx), xlog.WithErr(err))
					return
				}
				pinnedReplyUser = make(map[string]*userv1.UserInfo)
				pinnedReplyUser = userResp.Users
			})
		}

		resp, err := comment.Commenter().
			PageGetDetailedReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			replies = []*comment.DetailedReplyItem{}
		)

		if len(resp.Replies) > 0 {
			uidsMap := extractUidsMap(resp.Replies)

			// 发起请求获取uid的详细信息
			userResp, err := passport.Userer().
				BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}

			logx.Debugf("userResp = %v, uidsMap = %v", userResp.Users, uidsMap)
			// 拼接结果
			replies = make([]*comment.DetailedReplyItem, 0, len(resp.Replies))
			for _, item := range resp.Replies {
				details := comment.NewDetailedReplyItemFromPb(item, userResp.Users)
				replies = append(replies, details)
			}
		}

		var pinned *comment.DetailedReplyItem
		if req.Cursor == 0 {
			wg.Wait()
			// 置顶
			pinned = comment.NewDetailedReplyItemFromPb(pinnedReply, pinnedReplyUser)
		}

		httpx.OkJson(w, &comment.DetailedCommentRes{
			Replies:    replies,
			PinReply:   pinned,
			HasNext:    resp.HasNext,
			NextCursor: resp.NextCursor,
		})
	}
}

// 填入用户信息
func attachReplyUsers(ctx context.Context, replies []*commentv1.ReplyItem) ([]*comment.ReplyItem, error) {
	uidsMap := make(map[uint64]struct{})
	for _, root := range replies {
		uidsMap[root.Uid] = struct{}{}
	}

	userResp, err := passport.Userer().
		BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
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
		req, err := xhttp.ParseValidate[comment.DelReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = comment.Commenter().DelReply(r.Context(), &commentv1.DelReplyReq{ReplyId: req.ReplyId})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) PinComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.PinReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = comment.Commenter().PinReply(r.Context(), &commentv1.PinReplyReq{
			Oid:    req.Oid,
			Rid:    req.ReplyId,
			Action: commentv1.ReplyAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) LikeComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.ThumbUpReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = comment.Commenter().LikeAction(r.Context(), &commentv1.LikeActionReq{
			ReplyId: req.ReplyId,
			Action:  commentv1.ReplyAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.ThumbDownReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = comment.Commenter().DislikeAction(r.Context(), &commentv1.DislikeActionReq{
			ReplyId: req.ReplyId,
			Action:  commentv1.ReplyAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) GetCommentLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[comment.GetLikeCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := comment.Commenter().GetReplyLikeCount(r.Context(),
			&commentv1.GetReplyLikeCountReq{
				ReplyId: req.ReplyId,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &comment.GetLikeCountRes{
			ReplyId: res.ReplyId,
			Likes:   res.Count,
		})
	}
}
