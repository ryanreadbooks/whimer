package biz

import (
	bizcomment "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	biznote "github.com/ryanreadbooks/whimer/pilot/internal/biz/note"
	bizrelation "github.com/ryanreadbooks/whimer/pilot/internal/biz/relation"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
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
	NoteBiz      *biznote.Biz
}

func New(c *config.Config) *Biz {
	userBiz := bizuser.NewBiz(c)
	storageBiz := bizstorage.NewBiz(c)
	notifyBiz := bizsysnotify.NewBiz()
	return &Biz{
		UserBiz:      userBiz,
		FeedBiz:      bizfeed.NewBiz(),
		SearchBiz:    bizsearch.NewSearchBiz(c),
		CommentBiz:   bizcomment.NewBiz(),
		RelationBiz:  bizrelation.NewBiz(),
		SysNotifyBiz: notifyBiz,
		WhisperBiz:   bizwhisper.NewBiz(),
		UploadBiz:    storageBiz,
		NoteBiz:      biznote.NewBiz(storageBiz, notifyBiz, userBiz),
	}
}
