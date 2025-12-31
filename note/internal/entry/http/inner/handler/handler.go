package handler

import (
	"github.com/ryanreadbooks/whimer/note/internal/srv"
)

type Handler struct {
	Svc *srv.Service
}

func NewHandler(svc *srv.Service) *Handler {
	return &Handler{
		Svc: svc,
	}
}
