package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	
	"github.com/zeromicro/go-zero/rest/httpx"
)

func regFeedRoutes(group *xhttp.RouterGroup, svc *handler.Handler) {
	g := group.Group("/feed", middleware.CanLogin())
	{
		v1Group := g.Group("/v1")
		{
			v1Group.Get("/recommend", svc.Feed.GetRecommend())
			v1Group.Get("/detail", svc.Feed.GetNoteDetail(httpx.ParseForm))
			v1Group.Get("/note/:note_id", svc.Feed.GetNoteDetail(httpx.ParsePath))
			v1Group.Get("/notes/by_user", svc.Feed.GetNotesByUser())
		}
	}
}
