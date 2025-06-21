package backend

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/msg"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/profile"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/relation"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

type Handler struct {
	Config *config.Config

	Profile  *profile.Handler
	Comment  *comment.Handler
	Note     *note.Handler
	Relation *relation.Handler
	Chat     *msg.Handler
	User     *passport.UserHandler
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
	}

	return h
}
