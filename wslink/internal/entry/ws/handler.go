package ws

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	protov1 "github.com/ryanreadbooks/whimer/wslink/api/protocol/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/global"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
	protobuf "google.golang.org/protobuf/proto"
)

// 判断请求是否能升级
func (s *Server) isUpgradeAllowed(r *http.Request) error {
	if r == nil {
		return global.ErrBizInternal
	}

	// TODO

	return nil
}

// 协议升级成websocket
func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	// req param check
	reqId := r.URL.Query().Get("req_id")
	if reqId == "" {
		// req_id must exist
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(global.ErrReqIdMissing.Error()))
		return
	}
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// err != nil时 upgrader.Upgrade已经处理了
		return
	}

	conn := ws.CreateConnection(
		wsConn,
		ws.WithReadTimeout(s.conf.ReadTimeout.Duration()),
		ws.WithWriteTimeout(s.conf.WriteTimeout.Duration()),
	)
	conn.SetDevice(model.DeviceWeb)
	conn.SetLocalIp(config.GetIpAndPort())
	conn.SetReqId(reqId)

	ctx := r.Context()
	if err := s.OnCreate(ctx, conn); err != nil {
		// err is not nil, we deny this connection
		ws.RecoverConnection(conn)
		xhttp.Error(r, w, err)
		wsConn.Close()
		return
	}

	conn.SetOnData(s)
	conn.SetAfterClosed(s)

	concurrent.SafeGo(func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			ws.RecoverConnection(conn)
		}()

		conn.Loop(ctx)
	})
}

func (s *Server) OnCreate(ctx context.Context, conn *ws.Connection) error {
	return s.sessServ.OnCreate(ctx, conn)
}

func (s *Server) OnData(ctx context.Context, conn *ws.Connection, data []byte) error {
	// 数据上行
	var wire protov1.Protocol
	err := protobuf.Unmarshal(data, &wire)
	if err != nil {
		return s.sendError(ctx, conn, errUnexpectedProtocol)
	}

	// 上行只能是PING或者DATA
	switch wire.Meta.Flag {
	case protov1.Flag_FLAG_PING:
		if err := s.sessServ.Heatbeat(ctx, conn); err != nil {
			// heartbeat error we only log error here
			xlog.Msgf("ws server call heartbeat err").Err(err).Extras("cid", conn.GetId()).Errorx(ctx)
			return s.sendError(ctx, conn, err)
		}
		return s.sendPong(ctx, conn)
	case protov1.Flag_FLAG_DATA:
		return s.sessServ.OnData(ctx, conn, &wire)
	}

	return s.sendError(ctx, conn, errUnexpectedFlag)
}

func (s *Server) AfterClosed(ctx context.Context, id string) error {
	return s.sessServ.AfterClosed(ctx, id)
}

func (s *Server) sendWire(ctx context.Context, conn *ws.Connection, wire *protov1.Protocol) error {
	data, err := protobuf.Marshal(wire)
	if err != nil {
		xlog.Msgf("ws server protobuf marshal failed").Err(err).Errorx(ctx)
		return conn.Write([]byte("SERVER ERROR"))
	}

	return conn.Write(data)
}

func (s *Server) sendError(ctx context.Context, conn *ws.Connection, err error) error {
	wire := protov1.Protocol{
		Meta: &protov1.Meta{
			Flag: protov1.Flag_FLAG_ERR,
			Msg:  err.Error(),
		},
		Payload: nil,
	}

	return s.sendWire(ctx, conn, &wire)
}

func (s *Server) sendPong(ctx context.Context, conn *ws.Connection) error {
	wire := protov1.Protocol{
		Meta: &protov1.Meta{
			Flag: protov1.Flag_FLAG_PONG,
			Msg:  "PONG",
		},
	}

	return s.sendWire(ctx, conn, &wire)
}
