package handler

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/feed"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/msg"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/relation"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler/user"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
)

type Handler struct {
	Config *config.Config

	Comment  *comment.Handler
	Note     *note.Handler
	Relation *relation.Handler
	Chat     *msg.Handler
	User     *user.UserHandler
	Feed     *feed.Handler
}

func Init(c *config.Config) {
	infra.Init(c)
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	h := &Handler{
		Config:   c,
		Comment:  comment.NewHandler(c, bizz),
		Note:     note.NewHandler(c, bizz),
		Relation: relation.NewHandler(c, bizz),
		Chat:     msg.NewHandler(c),
		User:     user.NewUserHandler(c, bizz),
		Feed:     feed.NewHandler(c, bizz),
	}

	return h
}
