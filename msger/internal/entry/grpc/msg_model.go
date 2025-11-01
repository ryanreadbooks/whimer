package grpc

import (
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

func makePbMsg(m *bizp2p.ChatMsg) *pbmsg.Msg {
	if m == nil {
		return &pbmsg.Msg{}
	}

	return &pbmsg.Msg{
		MsgId:    m.MsgId,
		ChatId:   m.ChatId,
		Sender:   m.Sender,
		Receiver: m.Receiver,
		Status:   model.MsgStatusToPb(m.Status),
		Content: &pbmsg.MsgContent{
			Type: model.MsgTypeToPb(m.Type),
			Data: m.Content,
		},
		Seq: m.Seq,
	}
}
