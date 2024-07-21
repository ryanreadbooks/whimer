package handler

// import (
// 	"net/http"

// 	"github.com/ryanreadbooks/whimer/comment/internal/global"
// 	"github.com/ryanreadbooks/whimer/comment/internal/model"
// 	"github.com/ryanreadbooks/whimer/comment/internal/svc"
// 	"github.com/zeromicro/go-zero/rest/httpx"
// )

// // 发表评论
// func ReplyAdd(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		var req model.ReplyReq
// 		if err := httpx.ParseJsonBody(r, &req); err != nil {
// 			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
// 			return
// 		}

// 		if err := req.Validate(); err != nil {
// 			httpx.Error(w, err)
// 			return
// 		}

// 		err := c.CommentSvc.ReplyAdd(r.Context(), &req)
// 		if err != nil {
// 			httpx.Error(w, err)
// 			return
// 		}

// 		httpx.OkJson(w, nil)
// 	}
// }

// // 删除评论
// func ReplyDel(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 		err := c.CommentSvc.ReplyDel(r.Context())
// 		if err != nil {
// 			httpx.Error(w, err)
// 			return
// 		}

// 		httpx.OkJson(w, nil)
// 	}
// }

// // 点赞/取消点赞
// func LikeAction(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 	}
// }

// // 点踩/取消点踩
// func DislikeAction(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 	}
// }

// // 举报
// func ReportReply(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {

// 	}
// }

// // 置顶主评论
// func Pin(c *svc.ServiceContext) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {}
// }
