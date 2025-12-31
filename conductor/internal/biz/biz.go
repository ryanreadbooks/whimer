package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/shard"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra"
)

type Biz struct {
	rootCtx context.Context
	cancel  context.CancelFunc

	NamespaceBiz *NamespaceBiz
	TaskBiz      *TaskBiz
	WorkerBiz    *WorkerBiz
	ShardBiz     *shard.Biz
	CallbackBiz  *CallbackBiz
}

func NewBiz(rootCtx context.Context, c *config.Config) *Biz {
	return &Biz{
		rootCtx: rootCtx,

		NamespaceBiz: NewNamespaceBiz(infra.Dao().NamespaceDao),
		TaskBiz: NewTaskBiz(
			infra.Dao().TaskDao,
			infra.Dao().TaskHistoryDao,
		),
		WorkerBiz:   NewWorkerBiz(c),
		ShardBiz:    shard.NewBiz(c, infra.Etcd()),
		CallbackBiz: NewCallbackBiz(),
	}
}

func (b *Biz) Start() {
	b.ShardBiz.Run(b.rootCtx)
}

func (b *Biz) Stop() {
	b.ShardBiz.Stop()
}

func (b *Biz) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	return infra.Dao().DB().Transact(ctx, fn)
}
