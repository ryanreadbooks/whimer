package adapter

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/comment"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/storage"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
)

// 全局变量
var (
	noteCreatorAdapter  *note.CreatorAdapterImpl
	noteInteractAdapter *note.InteractAdapterImpl

	storageAdapter *storage.OssRepositoryImpl

	commentAdapter *comment.CommentAdapterImpl
)

func Init(c *config.Config) {
	noteCreatorAdapter = note.NewCreatorAdapterImpl(
		dep.NoteCreatorServer(),
		dep.NoteFeedServer(),
		dep.SearchServer(),
		dep.DocumentServer(),
	)
	noteInteractAdapter = note.NewInteractAdapterImpl(dep.NoteInteractServer())

	storageAdapter = storage.NewOssRepositoryImpl(
		storage.NewUploaders(c, dep.OssClient()),
		dep.OssClient(),
		dep.DisplayOssClient(),
	)

	commentAdapter = comment.NewCommentAdapterImpl(dep.Commenter())
}

func NoteCreatorAdapter() *note.CreatorAdapterImpl {
	return noteCreatorAdapter
}

func StorageAdapter() *storage.OssRepositoryImpl {
	return storageAdapter
}

func NoteInteractAdapter() *note.InteractAdapterImpl {
	return noteInteractAdapter
}

func CommentAdapter() *comment.CommentAdapterImpl {
	return commentAdapter
}
