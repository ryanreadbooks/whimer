package biz

import (
	bizcomment "github.com/ryanreadbooks/whimer/api-x/internal/biz/comment"
	bizfeed "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed"
	bizrelation "github.com/ryanreadbooks/whimer/api-x/internal/biz/relation"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	bizuser "github.com/ryanreadbooks/whimer/api-x/internal/biz/user"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
)

type Biz struct {
	FeedBiz     *bizfeed.Biz
	SearchBiz   *bizsearch.Biz
	UserBiz     *bizuser.Biz
	CommentBiz  *bizcomment.Biz
	RelationBiz *bizrelation.Biz
}

func New(c *config.Config) *Biz {
	return &Biz{
		FeedBiz:     bizfeed.NewFeedBiz(),
		SearchBiz:   bizsearch.NewSearchBiz(c),
		UserBiz:     bizuser.NewUserBiz(c),
		CommentBiz:  bizcomment.NewBiz(),
		RelationBiz: bizrelation.NewBiz(),
	}
}
