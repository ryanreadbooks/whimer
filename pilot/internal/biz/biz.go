package biz

import (
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizwhisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Biz struct {
	UserBiz      *bizuser.Biz
	SysNotifyBiz *bizsysnotify.Biz
	WhisperBiz   *bizwhisper.Biz
	UploadBiz    *bizstorage.Biz
}

func New(c *config.Config) *Biz {
	userBiz := bizuser.NewBiz(c)
	storageBiz := bizstorage.NewBiz(c)
	notifyBiz := bizsysnotify.NewBiz()
	return &Biz{
		UserBiz:      userBiz,
		SysNotifyBiz: notifyBiz,
		WhisperBiz:   bizwhisper.NewBiz(),
		UploadBiz:    storageBiz,
	}
}
