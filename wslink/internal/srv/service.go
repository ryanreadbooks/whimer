package srv

import (
	"context"
	"fmt"
	"strings"

	"github.com/reugn/async"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type Service struct {
	Config *config.Config

	// session连接
	sessions async.Map[string, ws.Session]
}

func NewService(c *config.Config) *Service {
	return &Service{
		Config:   c,
		sessions: async.NewShardedMap[string, ws.Session](128),
	}
}

func (s *Service) OnCreate(ctx context.Context, sess *ws.Session) error {
	xlog.Msgf("session %s is connected", sess.GetId()).Debugx(ctx)
	s.sessions.Put(sess.GetId(), sess)
	return nil
}

func (s *Service) OnData(ctx context.Context, sess *ws.Session, data []byte) error {
	fmt.Printf("data reached on %s, %s\n", sess.GetId(), data)
	// do echo
	return sess.WriteText(strings.ToUpper(string(data)))
}

func (s *Service) OnClosed(ctx context.Context, sid string) error {
	xlog.Msgf("session %s is closed", sid).Debugx(ctx)
	s.sessions.Remove(sid)
	return nil
}
