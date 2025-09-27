package biz

import (
	bizfeed "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	bizuser "github.com/ryanreadbooks/whimer/api-x/internal/biz/user"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

type Biz struct {
	FeedBiz   *bizfeed.FeedBiz
	SearchBiz *bizsearch.SearchBiz
	UserBiz   *bizuser.UserBiz
}

func New(c *config.Config) *Biz {
	return &Biz{
		FeedBiz:   bizfeed.NewFeedBiz(),
		SearchBiz: bizsearch.NewSearchBiz(c),
		UserBiz:   bizuser.NewUserBiz(c),
	}
}
