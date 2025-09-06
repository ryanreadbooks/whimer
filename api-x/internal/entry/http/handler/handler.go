package handler

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/feed"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/msg"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/passport"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/profile"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/relation"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
)

type Handler struct {
	Config *config.Config

	Profile  *profile.Handler
	Comment  *comment.Handler
	Note     *note.Handler
	Relation *relation.Handler
	Chat     *msg.Handler
	User     *passport.UserHandler
	Feed     *feed.Handler
}

func Init(c *config.Config) {
	infra.Init(c)
}

func NewHandler(c *config.Config) *Handler {
	h := &Handler{
		Config:   c,
		Profile:  profile.NewHandler(c),
		Comment:  comment.NewHandler(c),
		Note:     note.NewHandler(c),
		Relation: relation.NewHandler(c),
		Chat:     msg.NewHandler(c),
		User:     passport.NewUserHandler(c),
		Feed:     feed.NewHandler(c),
	}

	return h
}
