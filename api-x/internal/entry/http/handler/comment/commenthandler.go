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
func (h *Handler) PublishNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PubReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.Commenter().AddComment(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId := req.Oid.String()
		h.syncCommentCountToSearcher(ctx, noteId, 1)

		httpx.OkJson(w, &PubRes{CommentId: resp.CommentId})
	}
}

// 只获取主评论
func (h *Handler) PageGetNoteRootComments() http.HandlerFunc {
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
		rootReplies, err := infra.Commenter().PageGetComment(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 整合用户的信息
		var comments = []*CommentItem{}
		if len(rootReplies.Comments) > 0 {
			comments, err = attachReplyUsers(ctx, rootReplies.Comments)
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
		}

		attachCommentItemInteract(ctx, comments)

		httpx.OkJson(w, &CommentRes{
			Items:      comments,
			NextCursor: rootReplies.NextCursor,
			HasNext:    rootReplies.HasNext,
		})
	}
}

// 只获取子评论
func (h *Handler) PageGetNoteSubComments() http.HandlerFunc {
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
		subReplies, err := infra.Commenter().PageGetSubComment(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 填充评论的用户信息
		var comments = []*CommentItem{}
		if len(subReplies.Comments) > 0 {
			comments, err = attachReplyUsers(ctx, subReplies.Comments)
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
		}

		attachCommentItemInteract(ctx, comments)

		httpx.OkJson(w, &CommentRes{
			Items:      comments,
			HasNext:    subReplies.HasNext,
			NextCursor: subReplies.NextCursor,
		})
	}
}

func extractUidsMap(replies []*commentv1.DetailedCommentItem) map[int64]struct{} {
	uidsMap := make(map[int64]struct{})
	// 提取出主评论和子评论的uid
	for _, item := range replies {
		uidsMap[item.Root.Uid] = struct{}{}
		for _, sub := range item.SubComments.Items {
			uidsMap[sub.Uid] = struct{}{}
		}
	}

	return uidsMap
}

// 获取主评论信息（包含其下子评论）
func (h *Handler) PageGetNoteComments() http.HandlerFunc {
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
			pinnedReply     *commentv1.DetailedCommentItem
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
				resp, err := infra.Commenter().GetPinnedComment(ctx, &commentv1.GetPinnedCommentRequest{Oid: int64(req.Oid)})
				if err != nil {
					logx.Errorw("rpc get pin comment err", xlog.WithUid(ctx), xlog.WithErr(err))
					return
				}
				pinnedReply = resp.GetItem()
				if pinnedReply.GetRoot() == nil {
					// 可能不存在置顶评论
					return
				}

				userResp, err := infra.Userer().
					BatchGetUser(ctx,
						&userv1.BatchGetUserRequest{
							Uids: maps.Keys(extractUidsMap([]*commentv1.DetailedCommentItem{pinnedReply})),
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
			PageGetDetailedComment(ctx, req.AsDetailedPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			comments = []*DetailedCommentItem{}
		)

		if len(resp.Comments) > 0 {
			uidsMap := extractUidsMap(resp.Comments)

			// 发起请求获取uid的详细信息
			userResp, err := infra.Userer().
				BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}

			logx.Debugf("userResp = %v, uidsMap = %v", userResp.Users, uidsMap)
			// 拼接结果
			comments = make([]*DetailedCommentItem, 0, len(resp.Comments))
			for _, item := range resp.Comments {
				details := NewDetailedCommentItemFromPb(item, userResp.Users)
				comments = append(comments, details)
			}
		}

		var pinned *DetailedCommentItem
		wg.Wait()
		if req.Cursor == 0 {
			// 置顶
			if pinnedReply != nil && pinnedReply.GetRoot() != nil { // 有些可能没有设置置顶评论
				pinned = NewDetailedCommentItemFromPb(pinnedReply, pinnedReplyUser)
			}
		}

		temps := make([]*DetailedCommentItem, 0, len(comments)+1)
		temps = append(temps, comments...)
		if pinned != nil {
			temps = append(temps, pinned)
		}
		attachDetailCommentItemInteract(ctx, temps)

		httpx.OkJson(w, &DetailedCommentRes{
			Comments:   comments,
			PinComment: pinned,
			HasNext:    resp.HasNext,
			NextCursor: resp.NextCursor,
		})
	}
}

// 填入用户信息
func attachReplyUsers(ctx context.Context, replies []*commentv1.CommentItem) ([]*CommentItem, error) {
	uidsMap := make(map[int64]struct{})
	for _, root := range replies {
		uidsMap[root.Uid] = struct{}{}
	}

	userResp, err := infra.Userer().
		BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: maps.Keys(uidsMap)})
	if err != nil {
		return nil, err
	}

	res := make([]*CommentItem, 0, len(replies))
	for _, root := range replies {
		res = append(res, &CommentItem{
			CommentItemBase: NewCommentItemBaseFromPb(root),
			User:            userResp.Users[formatUid(root.Uid)],
		})
	}

	return res, nil
}

func formatUid(uid int64) string {
	return strconv.FormatInt(uid, 10)
}

func (h *Handler) DelNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[DelReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		_, err = infra.Commenter().DelComment(ctx, &commentv1.DelCommentRequest{
			CommentId: req.CommentId,
			Oid:       int64(req.Oid),
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

func (h *Handler) PinNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PinReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().PinComment(r.Context(), &commentv1.PinCommentRequest{
			Oid:       int64(req.Oid),
			CommentId: req.CommentId,
			Action:    commentv1.CommentAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) LikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ThumbUpReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().LikeAction(r.Context(), &commentv1.LikeActionRequest{
			CommentId: req.CommentId,
			Action:    commentv1.CommentAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ThumbDownReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_, err = infra.Commenter().DislikeAction(r.Context(), &commentv1.DislikeActionRequest{
			CommentId: req.CommentId,
			Action:    commentv1.CommentAction(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) GetNoteCommentLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetLikeCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := infra.Commenter().GetCommentLikeCount(r.Context(),
			&commentv1.GetCommentLikeCountRequest{
				CommentId: req.CommentId,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &GetLikeCountRes{
			Comment: res.CommentId,
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

func attachCommentItemInteract(ctx context.Context, items []*CommentItem) {
	if imodel.IsGuestFromCtx(ctx) {
		return
	}

	var uid = metadata.Uid(ctx)

	if len(items) == 0 {
		return
	}

	// collect all comment ids
	commentIds := make([]int64, 0, len(items))
	for _, item := range items {
		commentIds = append(commentIds, item.Id)
	}

	commentIds = xslice.Uniq(commentIds)

	// BatchCheckUserLikeReply有数量限制 此处需要分批处理
	var wg sync.WaitGroup
	var syncLikeStatus sync.Map
	err := xslice.BatchAsyncExec(&wg, commentIds, 50, func(start, end int) error {
		resp, err := infra.Commenter().BatchCheckUserLikeComment(ctx,
			&commentv1.BatchCheckUserLikeCommentRequest{
				Mappings: map[int64]*commentv1.BatchCheckUserLikeCommentRequest_CommentIdList{
					uid: {Ids: commentIds[start:end]},
				},
			})
		if err != nil {
			return err
		}

		if status, ok := resp.GetResults()[uid]; ok {
			for _, status := range status.List {
				syncLikeStatus.Store(status.GetCommentId(), status.GetLiked())
			}
		}

		return nil
	})

	if err != nil {
		xlog.Msg("comment handler failed to check user like comment status").Errorx(ctx)
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

func attachDetailCommentItemInteract(ctx context.Context, dItems []*DetailedCommentItem) {
	items := make([]*CommentItem, 0, len(dItems))
	for _, dItem := range dItems {
		items = append(items, dItem.Root)
		items = append(items, dItem.SubComments.Items...)
	}

	attachCommentItemInteract(ctx, items)
}

func (h *Handler) UploadCommentImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadImagesReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.Commenter().UploadCommentImages(ctx, &commentv1.UploadCommentImagesRequest{
			RequestedCount: req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
