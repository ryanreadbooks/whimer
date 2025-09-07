package biz

import (
	bizfeed "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

type Biz struct {
	FeedBiz   bizfeed.FeedBiz
	SearchBiz *bizsearch.SearchBiz
}

func New(c *config.Config) *Biz {
	return &Biz{
		FeedBiz:   bizfeed.NewFeedBiz(),
		SearchBiz: bizsearch.NewSearchBiz(c),
	}
}
