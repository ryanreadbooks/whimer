package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

func regSearchRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	searchV1 := group.Group("/search/v1", middleware.CanLogin())
	{
		// search note
		searchV1.Post("/note", h.Feed.SearchNotes())
		// get search notes available filters
		searchV1.Get("/note/filters", h.Feed.GetSearchNotesAvailableFilters())
	}
}
