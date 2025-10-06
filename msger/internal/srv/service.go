package srv

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
)

type Service struct {
	P2PChatSrv *P2PChatSrv
	SystemChatSrv *SystemChatSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{}
	// 基础设施初始化
	infra.Init(c)
	dep.Init(c)
	biz := biz.New()

	s.P2PChatSrv = NewP2PChatSrv(biz)
	s.SystemChatSrv = NewSystemChatSrv(biz)

	return s
}
