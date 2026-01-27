package adapter

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	domainpushcenter "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/comment"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note"
	adapterpushcenter "github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/relation"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/storage"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/systemnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/user"
	infracache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dao"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// 全局变量
var (
	noteCreatorAdapter  *note.CreatorAdapterImpl
	noteInteractAdapter *note.InteractAdapterImpl
	noteFeedAdapter     *note.NoteFeedAdapterImpl
	noteSearchAdapter   *note.NoteSearchAdapterImpl

	storageAdapter *storage.OssRepositoryImpl

	commentAdapter *comment.CommentAdapterImpl

	userSettingAdapter *user.UserSettingAdapter
	userAdapter        *user.UserAdapter

	relationAdapter *relation.RelationAdapterImpl

	systemNotifyAdapter *systemnotify.SystemNotifyAdapterImpl
)

func Init(c *config.Config, cache *redis.Redis) {
	noteCreatorAdapter = note.NewCreatorAdapterImpl(
		dep.NoteCreatorServer(),
		dep.NoteFeedServer(),
		dep.SearchServer(),
		dep.DocumentServer(),
	)
	noteInteractAdapter = note.NewInteractAdapterImpl(
		dep.NoteInteractServer(),
		infracache.NoteStatStore(),
	)
	noteFeedAdapter = note.NewNoteFeedAdapterImpl(
		dep.NoteFeedServer(),
		dep.NoteInteractServer(),
	)
	noteSearchAdapter = note.NewNoteSearchAdapterImpl(
		dep.DocumentServer(),
		dep.SearchServer(),
	)
	storageAdapter = storage.NewOssRepositoryImpl(
		storage.NewUploaders(c, dep.OssClient()),
		dep.OssClient(),
		dep.DisplayOssClient(),
	)

	commentAdapter = comment.NewCommentAdapterImpl(
		dep.Commenter(),
		infracache.NoteStatStore(),
	)

	userSettingAdapter = user.NewUserSettingAdapter(
		dao.Database().UserSettingDao,
		dep.RelationServer(),
	)
	userAdapter = user.NewUserAdapter(
		dep.Userer(),
	)

	relationAdapter = relation.NewRelationAdapterImpl(dep.RelationServer())

	systemNotifyAdapter = systemnotify.NewSystemNotifyAdapterImpl(
		dep.SystemNotifier(),
		dep.SystemChatter(),
	)

	domainpushcenter.SetPusher(adapterpushcenter.NewWsPusher())
}

func NoteCreatorAdapter() *note.CreatorAdapterImpl {
	return noteCreatorAdapter
}

func NoteFeedAdapter() *note.NoteFeedAdapterImpl {
	return noteFeedAdapter
}

func NoteInteractAdapter() *note.InteractAdapterImpl {
	return noteInteractAdapter
}

func NoteSearchAdapter() *note.NoteSearchAdapterImpl {
	return noteSearchAdapter
}

func StorageAdapter() *storage.OssRepositoryImpl {
	return storageAdapter
}

func CommentAdapter() *comment.CommentAdapterImpl {
	return commentAdapter
}

func UserSettingAdapter() *user.UserSettingAdapter {
	return userSettingAdapter
}

func UserAdapter() *user.UserAdapter {
	return userAdapter
}

func RelationAdapter() *relation.RelationAdapterImpl {
	return relationAdapter
}

func SystemNotifyAdapter() *systemnotify.SystemNotifyAdapterImpl {
	return systemNotifyAdapter
}
