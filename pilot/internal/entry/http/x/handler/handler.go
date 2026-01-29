package handler

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/comment"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/feed"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/msg"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/note"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/relation"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/upload"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler/user"
)

type Handler struct {
	Config *config.Config

	Comment  *comment.Handler
	Note     *note.Handler
	Relation *relation.Handler
	Chat     *msg.Handler
	User     *user.UserHandler
	Feed     *feed.Handler
	Upload   *upload.Handler
}

func NewHandler(c *config.Config, manager *app.Manager) *Handler {
	h := &Handler{
		Config:   c,
		Comment:  comment.NewHandler(c, manager),
		Note:     note.NewHandler(c, manager),
		Relation: relation.NewHandler(c, manager),
		Chat:     msg.NewHandler(c, manager),
		User:     user.NewUserHandler(c, manager),
		Feed:     feed.NewHandler(c, manager),
		Upload:   upload.NewHandler(c),
	}

	return h
}
