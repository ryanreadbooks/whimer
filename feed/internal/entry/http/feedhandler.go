package http

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/feed/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func feedRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedRecommendRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_ = req
	}
}

func feedDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}


// 获取笔记
// func (h *Handler) GetNote() http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParsePath, r)
// 		if err != nil {
// 			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
// 			return
// 		}

// 		var (
// 			uid    = metadata.Uid(r.Context())
// 			noteId = req.NoteId

// 			wg    sync.WaitGroup
// 			resp1 *commentv1.CountReplyRes
// 			resp2 *commentv1.CheckUserCommentOnObjectResponse
// 		)
// 		wg.Add(2)

// 		// 获取评论数
// 		concurrent.DoneInCtx(r.Context(), time.Second*10, func(ctx context.Context) {
// 			defer wg.Done()
// 			resp1, _ = comment.Commenter().CountReply(ctx, &commentv1.CountReplyReq{Oid: noteId})
// 		})

// 		concurrent.DoneInCtx(r.Context(), time.Second*10, func(ctx context.Context) {
// 			defer wg.Done()
// 			resp2, _ = comment.Commenter().CheckUserCommentOnObject(ctx, &commentv1.CheckUserCommentOnObjectRequest{
// 				Uid: uid,
// 				Oid: noteId,
// 			})
// 		})

// 		resp, err := note.NoteFeedServer().GetFeedNote(r.Context(), &notev1.GetFeedNoteRequest{
// 			NoteId: noteId,
// 		})

// 		if err != nil {
// 			xhttp.Error(r, w, err)
// 			return
// 		}

// 		wg.Wait()

// 		feed := note.NewFeedNoteItemFromPb(resp.Item)
// 		// 注入评论笔记的评论信息
// 		feed.Comments = resp1.GetNumReply()
// 		feed.Interact.Commented = resp2.GetCommented()

// 		httpx.OkJson(w, feed)
// 	}
// }