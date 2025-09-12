package srv

import (
	"github.com/ryanreadbooks/whimer/counter/internal/biz"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
)

type Service struct {
	CounterSrv *CounterSrv
	// CounterBiz *biz.CounterBiz
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{}
	bizz := biz.New(c)
	s.CounterSrv = NewCounterSrv(s, &bizz)
	// s.CounterBiz = bizz.CounterBiz

	return s
}
