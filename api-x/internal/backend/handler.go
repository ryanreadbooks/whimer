package backend

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

type Handler struct {
	Config *config.Config
}

func NewHandler(c *config.Config) *Handler {
	h := &Handler{
		Config: c,
	}
	passport.Init(c)
	note.Init(c)
	comment.Init(c)

	return h
}
