package biz

import (
	bizcomment "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	bizrelation "github.com/ryanreadbooks/whimer/pilot/internal/biz/relation"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Biz struct {
	FeedBiz      *bizfeed.Biz
	SearchBiz    *bizsearch.Biz
	UserBiz      *bizuser.Biz
	CommentBiz   *bizcomment.Biz
	RelationBiz  *bizrelation.Biz
	SysNotifyBiz *bizsysnotify.Biz
}

func New(c *config.Config) *Biz {
	return &Biz{
		FeedBiz:      bizfeed.NewFeedBiz(),
		SearchBiz:    bizsearch.NewSearchBiz(c),
		UserBiz:      bizuser.NewUserBiz(c),
		CommentBiz:   bizcomment.NewBiz(),
		RelationBiz:  bizrelation.NewBiz(),
		SysNotifyBiz: bizsysnotify.NewBiz(),
	}
}
