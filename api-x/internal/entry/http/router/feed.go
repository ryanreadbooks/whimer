package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

func regFeedRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	feedGroup := group.Group("/feed", middleware.CanLogin())
	{
		v1Group := feedGroup.Group("/v1")
		{
			v1Group.Get("/recommend", h.Feed.GetRecommend())
			v1Group.Get("/note/:note_id", h.Feed.GetNoteDetail())
			v1Group.Get("/notes/by_user", h.Feed.GetNotesByUser())
		}
	}
}
