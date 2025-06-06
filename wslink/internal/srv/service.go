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

	// 和本机建立的连接
	conns async.Map[string, ws.Connection]
}

func NewService(c *config.Config) *Service {
	return &Service{
		Config: c,
		conns:  async.NewShardedMap[string, ws.Connection](128),
	}
}

func (s *Service) OnCreate(ctx context.Context, sess *ws.Connection) error {
	xlog.Msgf("session %s is connected", sess.GetId()).Debugx(ctx)
	s.conns.Put(sess.GetId(), sess)
	return nil
}

func (s *Service) OnData(ctx context.Context, sess *ws.Connection, data []byte) error {
	fmt.Printf("data reached on %s, %s\n", sess.GetId(), data)
	// do echo
	return sess.WriteText(strings.ToUpper(string(data)))
}

func (s *Service) AfterClosed(ctx context.Context, sid string) error {
	xlog.Msgf("session %s is closed", sid).Debugx(ctx)
	s.conns.Remove(sid)
	return nil
}
