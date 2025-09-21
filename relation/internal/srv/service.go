package srv

import (
	"github.com/ryanreadbooks/whimer/relation/internal/biz"
	"github.com/ryanreadbooks/whimer/relation/internal/config"
)

type Service struct {
	Config *config.Config

	// domain service
	RelationSrv *RelationSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	ctx := &Service{
		Config: c,
	}

	// 业务初始化
	biz := biz.New()
	// 各个子service初始化

	ctx.RelationSrv = NewRelationSrv(ctx, biz)

	return ctx
}
