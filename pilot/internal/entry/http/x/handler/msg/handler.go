package msg

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	bizwhisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Handler struct {
	sysNotifyBiz *bizsysnotify.Biz
	userBiz      *bizuser.Biz
	whisperBiz   *bizwhisper.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		userBiz:      bizz.UserBiz,
		sysNotifyBiz: bizz.SysNotifyBiz,
		whisperBiz:   bizz.WhisperBiz,
	}
}
