package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	protov1 "github.com/ryanreadbooks/whimer/wslink/api/protocol/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"

	protobuf "google.golang.org/protobuf/proto"
)

// 封装ws.Connection并实现ISession接口
type ConnectionWrapper struct {
	*ws.Connection
}

// 关闭连接
func (cw *ConnectionWrapper) Close(ctx context.Context) {
	cw.GraceClose(ctx)
}

// 发送协议数据
func (cw *ConnectionWrapper) Send(ctx context.Context, data []byte) error {
	protocolData := protov1.Protocol{
		Meta: &protov1.Meta{
			Flag: protov1.Flag_FLAG_DATA,
		},
		Payload: data,
	}

	wireData, err := protobuf.Marshal(&protocolData)
	if err != nil {
		return xerror.Wrapf(err, "protobuf marshal failed").WithCtx(ctx)
	}

	return cw.Write(wireData)
}

type PushLocalConnReq struct {
	Conn biz.Session
	Data []byte
}

type PushNonLocalConnReq struct {
	Conn       biz.UnSendableSession
	Data       []byte
	ForwardCnt int
}

func FormatPushLocalConnReq(locals []biz.Session, device model.Device, data []byte) []*PushLocalConnReq {
	localTarges := make([]*PushLocalConnReq, 0, len(locals))
	for _, l := range locals {
		if device != "" && l.GetDevice() != device {
			continue
		}

		localTarges = append(localTarges, &PushLocalConnReq{
			Conn: l,
			Data: data,
		})
	}

	return localTarges
}

func FormatPushNonLocalConnReq(nonLocals []biz.UnSendableSession, device model.Device, data []byte) []*PushNonLocalConnReq {
	nonLocalTargets := make([]*PushNonLocalConnReq, 0, len(nonLocals))
	for _, nl := range nonLocals {
		if device != "" && nl.GetDevice() != device {
			continue
		}

		nonLocalTargets = append(nonLocalTargets, &PushNonLocalConnReq{
			Conn:       nl,
			Data:       data,
			ForwardCnt: 1, // 第一次转发
		})

	}
	return nonLocalTargets
}
