package handler

// import (
// 	"github.com/ryanreadbooks/whimer/comment/internal/external"
// 	"github.com/ryanreadbooks/whimer/comment/internal/svc"
// 	"github.com/ryanreadbooks/whimer/misc/xhttp"
// 	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
// 	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/remote"

// 	"github.com/zeromicro/go-zero/core/service"
// 	"github.com/zeromicro/go-zero/rest"
// )

// func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
// 	xGroup := xhttp.NewRouterGroup(engine)
	// regReplyRoutes(xGroup, ctx)

// 	mod := ctx.Config.Http.Mode
// 	if mod == service.DevMode || mod == service.TestMode {
// 		engine.PrintRoutes()
// 	}
// }

// 评论相关路由
// func regReplyRoutes(group *xhttp.RouterGroup, c *svc.ServiceContext) {
// 	replyGroup := group.Group("/reply",
// 		auth.User(external.GetAuther()),
// 		remote.Addr,
// 	)

// 	{
// 		v1 := replyGroup.Group("/v1")
// 		{
// 			v1.Post("/add", ReplyAdd(c))
// 			v1.Post("/del", ReplyDel(c))
// 			v1.Post("/like", LikeAction(c))
// 			v1.Post("/dislike", DislikeAction(c))
// 			v1.Post("/report", ReportReply(c))
// 			v1.Post("/pin", Pin(c))
// 		}
// 	}
// }
