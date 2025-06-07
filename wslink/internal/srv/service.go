package srv

import (
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
)

type Service struct {
	c *config.Config

	// domain services
	SessionService *SessionService
	PushService    *PushService
}

func New(c *config.Config) *Service {
	s := &Service{
		c: c,
	}

	b := biz.New()
	s.SessionService = NewSessionService(c, b)
	s.PushService = NewPushService(b)

	return s
}
