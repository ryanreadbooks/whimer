package biz

import (
	bizcomment "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	bizrelation "github.com/ryanreadbooks/whimer/pilot/internal/biz/relation"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/storage"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/user"
	bizwhisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Biz struct {
	FeedBiz      *bizfeed.Biz
	SearchBiz    *bizsearch.Biz
	UserBiz      *bizuser.Biz
	CommentBiz   *bizcomment.Biz
	RelationBiz  *bizrelation.Biz
	SysNotifyBiz *bizsysnotify.Biz
	WhisperBiz   *bizwhisper.Biz
	UploadBiz    *bizstorage.Biz
}

func New(c *config.Config) *Biz {
	userBiz := bizuser.NewBiz(c)
	return &Biz{
		UserBiz:      userBiz,
		FeedBiz:      bizfeed.NewBiz(),
		SearchBiz:    bizsearch.NewSearchBiz(c),
		CommentBiz:   bizcomment.NewBiz(),
		RelationBiz:  bizrelation.NewBiz(),
		SysNotifyBiz: bizsysnotify.NewBiz(),
		WhisperBiz:   bizwhisper.NewBiz(),
		UploadBiz:    bizstorage.NewBiz(c),
	}
}
