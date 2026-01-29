package msg

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	sysnotifyapp "github.com/ryanreadbooks/whimer/pilot/internal/app/systemnotify"
	whisperapp "github.com/ryanreadbooks/whimer/pilot/internal/app/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Handler struct {
	sysNotifyApp *sysnotifyapp.Service
	whisperApp   *whisperapp.Service
}

func NewHandler(c *config.Config, appMgr *app.Manager) *Handler {
	return &Handler{
		sysNotifyApp: appMgr.SystemNotifyApp,
		whisperApp:   appMgr.WhisperApp,
	}
}
