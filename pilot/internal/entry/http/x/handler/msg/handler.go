package msg

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	sysnotifyapp "github.com/ryanreadbooks/whimer/pilot/internal/app/systemnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizwhisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
	adapteruser "github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/user"
)

type Handler struct {
	sysNotifyApp *sysnotifyapp.Service
	whisperBiz   *bizwhisper.Biz
	userAdapter  *adapteruser.UserAdapter
}

func NewHandler(c *config.Config, bizz *biz.Biz, appMgr *app.Manager) *Handler {
	return &Handler{
		sysNotifyApp: appMgr.SystemNotifyApp,
		whisperBiz:   bizz.WhisperBiz,
		userAdapter:  adapter.UserAdapter(),
	}
}
