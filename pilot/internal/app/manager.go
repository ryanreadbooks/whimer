package app

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteevent"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/relation"
	sysnotifyapp "github.com/ryanreadbooks/whimer/pilot/internal/app/systemnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify"
	userdomain "github.com/ryanreadbooks/whimer/pilot/internal/domain/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/repo"
)

type Manager struct {
	NoteCreatorApp  *notecreator.Service
	NoteInteractApp *noteinteract.Service
	NoteFeedApp     *notefeed.Service
	NoteEventApp    *noteevent.Service
	RelationApp     *relation.Service
	UserApp         *user.Service
	CommentApp      *comment.Service
	SystemNotifyApp *sysnotifyapp.Service
}

func NewManager(c *config.Config) *Manager {
	userDomainService := userdomain.NewDomainService(
		adapter.UserAdapter(),
		repo.RecentContactRepo(),
	)
	systemNotifyDomainService := systemnotify.NewDomainService(
		adapter.SystemNotifyAdapter(),
	)

	m := &Manager{
		NoteCreatorApp: notecreator.NewService(
			adapter.NoteCreatorAdapter(),
			adapter.NoteInteractAdapter(),
			adapter.CommentAdapter(),
			adapter.StorageAdapter(),
		),
		NoteInteractApp: noteinteract.NewService(
			adapter.NoteInteractAdapter(),
			adapter.NoteFeedAdapter(),
			adapter.CommentAdapter(),
			systemNotifyDomainService,
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
			adapter.NoteFeedAdapter(),
			adapter.NoteSearchAdapter(),
			adapter.UserAdapter(),
			systemNotifyDomainService,
			userDomainService,
		),
		RelationApp: relation.NewService(
			adapter.RelationAdapter(),
		),
		UserApp: user.NewService(
			userDomainService,
			adapter.UserAdapter(),
			adapter.RelationAdapter(),
			adapter.NoteCreatorAdapter(),
			adapter.NoteFeedAdapter(),
			adapter.UserSettingAdapter(),
			repo.RecentContactRepo(),
		),
		CommentApp: comment.NewService(
			adapter.CommentAdapter(),
			userDomainService,
			adapter.UserAdapter(),
			adapter.NoteFeedAdapter(),
			adapter.NoteCreatorAdapter(),
			adapter.StorageAdapter(),
			systemNotifyDomainService,
		),
		SystemNotifyApp: sysnotifyapp.NewService(
			systemNotifyDomainService,
			adapter.NoteFeedAdapter(),
			adapter.NoteInteractAdapter(),
			adapter.CommentAdapter(),
			adapter.UserAdapter(),
		),
	}

	return m
}
