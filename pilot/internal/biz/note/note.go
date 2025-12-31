package note

import (
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
)

type Biz struct {
	storageBiz *bizstorage.Biz
	notifyBiz  *bizsysnotify.Biz
	userBiz    *bizuser.Biz
}

func NewBiz(
	storageBiz *bizstorage.Biz,
	notifyBiz *bizsysnotify.Biz,
	userBiz *bizuser.Biz) *Biz {
	return &Biz{
		storageBiz: storageBiz,
		notifyBiz:  notifyBiz,
		userBiz:    userBiz,
	}
}
