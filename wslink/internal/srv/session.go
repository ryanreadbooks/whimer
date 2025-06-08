package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	v1 "github.com/ryanreadbooks/whimer/wslink/api/protocol/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type SessionService struct {
	c          *config.Config
	sessionBiz biz.SessionBiz
}

func NewSessionService(c *config.Config, b biz.Biz) *SessionService {
	return &SessionService{
		c:          c,
		sessionBiz: b.SessionBiz,
	}
}

// 连接已经创建
func (s *SessionService) OnCreate(ctx context.Context, conn *ws.Connection) error {
	var uid = metadata.Uid(ctx)
	if err := s.sessionBiz.Connect(ctx, &ConnectionWrapper{
		Connection: conn,
	}); err != nil {
		return xerror.Wrapf(err, "failed to create connection").WithCtx(ctx)
	}
	xlog.Msgf("connection %s is connected, uid: %d", conn.GetId(), uid).Debugx(ctx)

	return nil
}

// 数据上行
func (s *SessionService) OnData(ctx context.Context, conn *ws.Connection, wire *v1.Protocol) error {
	xlog.Msgf("connection %s data reached: %s", conn.GetId(), wire.GetPayload()).Debugx(ctx)
	return nil
}

// 连接已经关闭
func (s *SessionService) AfterClosed(ctx context.Context, cid string) error {
	err := s.sessionBiz.Disconnect(ctx, cid)
	if err != nil {
		xlog.Msgf("failed to after close connection").Extras("cid", cid).Errorx(ctx)
	}

	xlog.Msgf("connection %s is closed", cid).Debugx(ctx)

	return nil
}

// graceful close action before system quit
func (s *SessionService) Close(ctx context.Context) {
	s.sessionBiz.Close(ctx)
}

func (s *SessionService) Heatbeat(ctx context.Context, conn *ws.Connection) error {
	xlog.Msgf("connection %s heartbeat", conn.GetId()).Debugx(ctx)
	return s.sessionBiz.Heartbeat(ctx, &ConnectionWrapper{
		Connection: conn,
	})
}
