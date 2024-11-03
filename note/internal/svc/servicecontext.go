package svc

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
)

type ServiceContext struct {
	Config *config.Config

	// utilities
	OssKeyGen *keygen.Generator

	// domain service
	NoteAdminSvc    *NoteAdminSvc
	NoteFeedSvc     *NoteFeedSvc
	NoteInteractSvc *NoteInteractSvc
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	// 基础设施初始化
	infra.Init(c)

	// utilities
	ctx.OssKeyGen = keygen.NewGenerator(
		keygen.WithBucket(c.Oss.Bucket),
		keygen.WithPrefix(c.Oss.Prefix),
		keygen.WithPrependBucket(true),
	)

	// 各个子service初始化
	ctx.NoteAdminSvc = NewNoteAdminSvc(ctx)
	ctx.NoteFeedSvc = NewNoteFeedSvc(ctx)
	ctx.NoteInteractSvc = NewNoteInteractSvc(ctx)

	return ctx
}
