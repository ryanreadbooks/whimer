package chat

import "github.com/ryanreadbooks/whimer/api-x/internal/config"

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}
