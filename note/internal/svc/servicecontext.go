package svc

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/infra/repo"
	"github.com/zeromicro/go-zero/core/stores/redis"
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
	dao := repo.New(c)
	ctx := &ServiceContext{
		Config: c,
	}

	// 外部依赖客户端初始化
	infra.Init(c)

	// utilities
	ctx.OssKeyGen = keygen.NewGenerator(
		keygen.WithBucket(c.Oss.Bucket),
		keygen.WithPrefix(c.Oss.Prefix),
		keygen.WithPrependBucket(true),
	)

	cache := redis.MustNewRedis(c.Redis)

	// 各个子service初始化
	ctx.NoteAdminSvc = NewNoteAdminSvc(ctx, dao, cache)
	ctx.NoteFeedSvc = NewNoteFeedSvc(ctx)
	ctx.NoteInteractSvc = NewNoteInteractSvc(ctx)

	return ctx
}
