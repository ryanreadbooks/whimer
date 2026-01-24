package app

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
)

type Manager struct {
	NoteCreatorApp  *notecreator.Service
	NoteInteractApp *noteinteract.Service
}

func NewManager(c *config.Config) *Manager {
	return &Manager{
		NoteCreatorApp: notecreator.NewService(
			adapter.NoteCreatorAdapter(),
			adapter.NoteInteractAdapter(),
			adapter.CommentAdapter(),
			adapter.StorageAdapter(),
		),
		NoteInteractApp: noteinteract.NewService(
			adapter.NoteInteractAdapter(),
			adapter.CommentAdapter(),
		),
	}
}
