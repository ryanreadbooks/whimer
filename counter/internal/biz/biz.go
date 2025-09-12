package biz

import "github.com/ryanreadbooks/whimer/counter/internal/config"

type Biz struct {
	CounterBiz *CounterBiz
}

func New(c *config.Config) Biz {
	b := Biz{}
	b.CounterBiz = MustNewCounterBiz(c)
	return b
}
