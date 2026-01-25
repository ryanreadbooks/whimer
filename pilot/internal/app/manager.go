package app

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteevent"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/relation"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
)

type Manager struct {
	NoteCreatorApp  *notecreator.Service
	NoteInteractApp *noteinteract.Service
	NoteFeedApp     *notefeed.Service
	NoteEventApp    *noteevent.Service
	RelationApp     *relation.Service
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
		NoteFeedApp: notefeed.NewService(
			adapter.NoteFeedAdapter(),
			adapter.NoteInteractAdapter(),
			adapter.NoteSearchAdapter(),
			adapter.UserAdapter(),
			adapter.RelationAdapter(),
			adapter.StorageAdapter(),
			adapter.CommentAdapter(),
			adapter.UserSettingAdapter(),
		),
		NoteEventApp: noteevent.NewService(
			adapter.NoteSearchAdapter(),
			adapter.UserAdapter(),
		),
		RelationApp: relation.NewService(
			adapter.RelationAdapter(),
		),
	}
}
