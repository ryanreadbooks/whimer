package srv

import (
	"context"
	"fmt"
	"strings"

	"github.com/reugn/async"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type Service struct {
	Config *config.Config
	bizz   biz.Biz

	// 和本机建立的连接
	conns async.Map[string, ws.Connection]
}

func NewService(c *config.Config) *Service {
	b := biz.New()
	return &Service{
		Config: c,
		bizz:   b,
		conns:  async.NewShardedMap[string, ws.Connection](128),
	}
}

// 连接已经创建
func (s *Service) OnCreate(ctx context.Context, conn *ws.Connection) error {
	var uid = metadata.Uid(ctx)
	if err := s.bizz.SessionBiz.Connect(ctx, conn); err != nil {
		return xerror.Wrapf(err, "failed to create connection").WithCtx(ctx)
	}
	xlog.Msgf("connection %s is connected, uid: %d", conn.GetId(), uid).Debugx(ctx)
	s.conns.Put(conn.GetId(), conn)

	return nil
}

// 数据上行
func (s *Service) OnData(ctx context.Context, conn *ws.Connection, data []byte) error {
	fmt.Printf("data reached on %s, %s\n", conn.GetId(), data)
	// do echo
	return conn.WriteText(strings.ToUpper(string(data)))
}

// 连接已经关闭
func (s *Service) AfterClosed(ctx context.Context, cid string) error {
	err := s.bizz.SessionBiz.Disconnect(ctx, cid)
	if err != nil {
		xlog.Msgf("failed to after close connection").Extras("cid", cid).Errorx(ctx)
	}

	xlog.Msgf("connection %s is closed", cid).Debugx(ctx)
	s.conns.Remove(cid)

	return nil
}

// graceful close action before system quit
func (s *Service) Close(ctx context.Context) {
	keys := s.conns.KeySet()
	s.bizz.SessionBiz.OfflineSession(ctx, keys)
	s.conns.Clear()
}
