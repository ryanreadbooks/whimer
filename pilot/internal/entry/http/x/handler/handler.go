package handler

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
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

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	h := &Handler{
		Config:   c,
		Comment:  comment.NewHandler(c, bizz),
		Note:     note.NewHandler(c, bizz),
		Relation: relation.NewHandler(c, bizz),
		Chat:     msg.NewHandler(c, bizz),
		User:     user.NewUserHandler(c, bizz),
		Feed:     feed.NewHandler(c, bizz),
		Upload:   upload.NewHandler(c, bizz),
	}

	return h
}
