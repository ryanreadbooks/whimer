package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/note/internal/entry/http/inner/handler"
)

func regNoteProcessRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	noteProcessGroup := group.Group("/note_process")
	{
		// /inner/api/v1/dev/note_process/callback
		noteProcessGroup.Post("/callback", h.NoteProcessCallback())
	}
}
