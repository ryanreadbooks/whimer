package msg

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Handler struct {
	sysNotifyBiz *bizsysnotify.Biz
	userBiz      *bizuser.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		userBiz:      bizz.UserBiz,
		sysNotifyBiz: bizz.SysNotifyBiz,
	}
}


