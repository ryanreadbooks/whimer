package comment

import (
	"context"
	"net/http"
	"strconv"
	"sync"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	imodel "github.com/ryanreadbooks/whimer/api-x/internal/model"
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	searchBiz *bizsearch.SearchBiz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		searchBiz: bizz.SearchBiz,
	}
}

func (h *Handler) syncCommentCountToSearcher(ctx context.Context, noteId string, incr int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "note.handler.commentnote.synces",
		Job: func(ctx context.Context) error {
			err := h.searchBiz.NoteStatSyncer.AddCommentCount(ctx, noteId, incr)
			if err != nil {
				xlog.Msg("note stat add comment count failed").
					Extras("note_id", noteId, "incr", incr).
					Err(err).Errorx(ctx)
			}

			return err
		},
	})
}

// 发表评论
func (h *Handler) PublishComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PubReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.Commenter().AddReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId := req.Oid.String()
		h.syncCommentCountToSearcher(ctx, noteId, 1)

		httpx.OkJson(w, &PubRes{ReplyId: resp.ReplyId})
	}
}

// 只获取主评论
func (h *Handler) PageGetRoots() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		rootReplies, err := infra.Commenter().
			PageGetReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 整合用户的信息
		var replies = []*ReplyItem{}
		if len(rootReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, rootReplies.Replies)
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
		}

		attachReplyItemInteract(ctx, replies)

		httpx.OkJson(w, &CommentRes{
			Replies:    replies,
			NextCursor: rootReplies.NextCursor,
			HasNext:    rootReplies.HasNext,
		})
	}
}

// 只获取子评论
func (h *Handler) PageGetSubs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetSubCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		subReplies, err := infra.Commenter().
			PageGetSubReply(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 填充评论的用户信息
		var replies = []*ReplyItem{}
		if len(subReplies.Replies) > 0 {
			replies, err = attachReplyUsers(ctx, subReplies.Replies)
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
		}

		attachReplyItemInteract(ctx, replies)

		httpx.OkJson(w, &CommentRes{
			Replies:    replies,
			HasNext:    subReplies.HasNext,
			NextCursor: subReplies.NextCursor,
		})
	}
}

func extractUidsMap(replies []*commentv1.DetailedReplyItem) map[int64]struct{} {
	uidsMap := make(map[int64]struct{})
	// 提取出主评论和子评论的uid
	for _, item := range replies {
		uidsMap[item.Root.Uid] = struct{}{}
		for _, sub := range item.SubReplies.Items {
			uidsMap[sub.Uid] = struct{}{}
		}
	}

	return uidsMap
}

// 获取主评论信息（包含其下子评论）
func (h *Handler) PageGetReplies() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
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
				resp, err := infra.Commenter().
					GetPinnedReply(ctx, &commentv1.GetPinnedReplyRequest{Oid: int64(req.Oid)})
				if err != nil {
					logx.Errorw("rpc get pin reply err", xlog.WithUid(ctx), xlog.WithErr(err))
					return
				}
				pinnedReply = resp.GetReply()
				if pinnedReply.GetRoot() == nil {
					// 可能不存在置顶评论
					return
				}

				userResp, err := infra.Userer().
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

		resp, err := infra.Commenter().
			PageGetDetailedReply(ctx, req.AsDetailedPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			replies = []*DetailedReplyItem{}
		)

		if len(resp.Replies) > 0 {
			uidsMap := extractUidsMap(resp.Replies)

			// 发起请求获取uid的详细信息
			userResp, err := infra.Userer().
				BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}

			logx.Debugf("userResp = %v, uidsMap = %v", userResp.Users, uidsMap)
			// 拼接结果
			replies = make([]*DetailedReplyItem, 0, len(resp.Replies))
			for _, item := range resp.Replies {
				details := NewDetailedReplyItemFromPb(item, userResp.Users)
				replies = append(replies, details)
			}
		}

		var pinned *DetailedReplyItem
		wg.Wait()
		if req.Cursor == 0 {
			// 置顶
			if pinnedReply != nil && pinnedReply.GetRoot() != nil { // 有些可能没有设置置顶评论
				pinned = NewDetailedReplyItemFromPb(pinnedReply, pinnedReplyUser)
			}
		}

		temps := make([]*DetailedReplyItem, 0, len(replies)+1)
		temps = append(temps, replies...)
		if pinned != nil {
			temps = append(temps, pinned)
		}
		attachDetailReplyItemInteract(ctx, temps)

		httpx.OkJson(w, &DetailedCommentRes{
			Replies:    replies,
			PinReply:   pinned,
			HasNext:    resp.HasNext,
			NextCursor: resp.NextCursor,
		})
	}
}

// 填入用户信息
func attachReplyUsers(ctx context.Context, replies []*commentv1.ReplyItem) ([]*ReplyItem, error) {
	uidsMap := make(map[int64]struct{})
	for _, root := range replies {
		uidsMap[root.Uid] = struct{}{}
	}

	userResp, err := infra.Userer().
		BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
	if err != nil {
		return nil, err
	}

	res := make([]*ReplyItem, 0, len(replies))
	for _, root := range replies {
		res = append(res, &ReplyItem{
			ReplyItemBase: NewReplyItemBaseFromPb(root),
			User:          userResp.Users[formatUid(root.Uid)],
		})
	}

	return res, nil
}

func formatUid(uid int64) string {
	return strconv.FormatInt(uid, 10)
}

func (h *Handler) DelComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[DelReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		_, err = infra.Commenter().DelReply(ctx, &commentv1.DelReplyRequest{
			ReplyId: req.ReplyId,
			Oid:     int64(req.Oid),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId := req.Oid.String()
		h.syncCommentCountToSearcher(ctx, noteId, -1)

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) PinComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PinReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().PinReply(r.Context(), &commentv1.PinReplyRequest{
			Oid:    int64(req.Oid),
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
		req, err := xhttp.ParseValidate[ThumbUpReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().LikeAction(r.Context(), &commentv1.LikeActionRequest{
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
		req, err := xhttp.ParseValidate[ThumbDownReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().DislikeAction(r.Context(), &commentv1.DislikeActionRequest{
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
		req, err := xhttp.ParseValidate[GetLikeCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := infra.Commenter().GetReplyLikeCount(r.Context(),
			&commentv1.GetReplyLikeCountRequest{
				ReplyId: req.ReplyId,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &GetLikeCountRes{
			ReplyId: res.ReplyId,
			Likes:   res.Count,
		})
	}
}

func (h *Handler) checkHasNote(ctx context.Context, noteId int64) error {
	if resp, err := infra.NoteCreatorServer().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: noteId,
		}); err != nil {
		return err
	} else {
		if !resp.Exist {
			return xerror.ErrArgs.Msg("笔记不存在")
		}
	}

	return nil
}

func attachReplyItemInteract(ctx context.Context, items []*ReplyItem) {
	if imodel.IsGuestFromCtx(ctx) {
		return
	}

	var uid = metadata.Uid(ctx)

	if len(items) == 0 {
		return
	}

	// collect all reply ids
	replyIds := make([]int64, 0, len(items))
	for _, item := range items {
		replyIds = append(replyIds, item.Id)
	}

	replyIds = xslice.Uniq(replyIds)

	// BatchCheckUserLikeReply有数量限制 此处需要分批处理
	var wg sync.WaitGroup
	var syncLikeStatus sync.Map
	err := xslice.BatchAsyncExec(&wg, replyIds, 50, func(start, end int) error {
		resp, err := infra.Commenter().BatchCheckUserLikeReply(ctx,
			&commentv1.BatchCheckUserLikeReplyRequest{
				Mappings: map[int64]*commentv1.BatchCheckUserLikeReplyRequest_ReplyIdList{
					uid: {Ids: replyIds[start:end]},
				},
			})
		if err != nil {
			return err
		}

		if status, ok := resp.GetResults()[uid]; ok {
			for _, status := range status.List {
				syncLikeStatus.Store(status.GetReplyId(), status.GetLiked())
			}
		}

		return nil
	})

	if err != nil {
		xlog.Msg("comment handler failed to check user like reply status").Errorx(ctx)
		return
	}

	// fill items
	for _, item := range items {
		if v, ok := syncLikeStatus.Load(item.Id); ok {
			if vv, yes := v.(bool); yes {
				item.Interact.Liked = vv
			}
		}
	}
}

func attachDetailReplyItemInteract(ctx context.Context, dItems []*DetailedReplyItem) {
	items := make([]*ReplyItem, 0, len(dItems))
	for _, dItem := range dItems {
		items = append(items, dItem.Root)
		items = append(items, dItem.SubReplies.Items...)
	}

	attachReplyItemInteract(ctx, items)
}
