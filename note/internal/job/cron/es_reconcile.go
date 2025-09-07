package job

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
)

// TODO 数据库和es之间的定时同步任务
type EsReconciler struct {
	bizz biz.Biz

	ctx    context.Context
	cancel context.CancelFunc
}

type EsReconcileConfig struct {
}

func NewEsReconciler(biz biz.Biz, cfg EsReconcileConfig) *EsReconciler {
	ctx, cancel := context.WithCancel(context.Background())
	r := &EsReconciler{
		bizz:   biz,
		ctx:    ctx,
		cancel: cancel,
	}

	return r
}

func (r *EsReconciler) Start() {

}

func (r *EsReconciler) Stop() {
	r.cancel()
}
