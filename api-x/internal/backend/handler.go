package backend

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/chat"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
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
	Chat     *chat.Handler
}

func Init(c *config.Config) {
	passport.Init(c)
	note.Init(c)
	comment.Init(c)
	relation.Init(c)
	chat.Init(c)
}

func NewHandler(c *config.Config) *Handler {
	h := &Handler{
		Config:   c,
		Profile:  profile.NewHandler(c),
		Comment:  comment.NewHandler(c),
		Note:     note.NewHandler(c),
		Relation: relation.NewHandler(c),
		Chat:     chat.NewHandler(c),
	}

	return h
}
