package biz

import (
	bizwhisper "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

type Biz struct {
	WhisperBiz *bizwhisper.Biz
}

func New(c *config.Config) *Biz {
	return &Biz{
		WhisperBiz: bizwhisper.NewBiz(),
	}
}
