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
	ShardBiz     *shard.Biz
}

func NewBiz(c *config.Config) *Biz {
	rootCtx, cancel := context.WithCancel(context.Background())
	return &Biz{
		rootCtx: rootCtx,
		cancel:  cancel,

		NamespaceBiz: NewNamespaceBiz(infra.Dao().NamespaceDao),
		TaskBiz: NewTaskBiz(
			infra.Dao().TaskDao,
			infra.Dao().TaskHistoryDao,
			infra.Dao().LockDao,
		),
		ShardBiz: shard.NewBiz(c, infra.Etcd()),
	}
}

func (b *Biz) Start() {
	b.ShardBiz.Run(b.rootCtx)
}

func (b *Biz) Stop() {
	b.cancel()
	b.ShardBiz.Stop()
}
