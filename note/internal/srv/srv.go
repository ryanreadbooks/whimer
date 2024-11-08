package srv

import (
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
)

type ServiceContext struct {
	Config *config.Config

	// domain service
	NoteCreatorSrv  *NoteCreatorSrv
	NoteFeedSrv     *NoteFeedSrv
	NoteInteractSrv *NoteInteractSrv
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	ctx := &ServiceContext{
		Config: c,
	}

	// 基础设施初始化
	infra.Init(c)
	// 业务初始化
	biz := biz.New()
	// 各个子service初始化
	ctx.NoteCreatorSrv = NewNoteCreatorSrv(ctx, biz)
	ctx.NoteFeedSrv = NewNoteFeedSrv(ctx, biz)
	ctx.NoteInteractSrv = NewNoteInteractSrv(ctx, biz)

	return ctx
}
