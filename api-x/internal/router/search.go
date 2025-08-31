package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
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