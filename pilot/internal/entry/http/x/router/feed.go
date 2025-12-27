package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

func regFeedRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	feedGroup := group.Group("/feed", middleware.CanLogin())
	{
		v1Group := feedGroup.Group("/v1")
		{
			v1Group.Get("/recommend", h.Feed.GetRecommend())
			v1Group.Get("/note/:note_id", h.Feed.GetNoteDetail())
			v1Group.Get("/note/by_user", h.Feed.GetNotesByUser())
		}
	}
}
