package handler

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Handler struct {
	Config *config.Config
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	h := &Handler{
		Config: c,
	}

	return h
}
