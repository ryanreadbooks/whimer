package srv

import (
	"github.com/ryanreadbooks/whimer/feed/internal/biz"
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/feed/internal/infra"
)

// globals
var (
	Service *service
)

type service struct {
	FeedBiz biz.FeedBiz
}

func Init(c *config.Config) {
	infra.Init(c)
	Service = &service{
		FeedBiz: biz.NewFeedBiz(),
	}
}
