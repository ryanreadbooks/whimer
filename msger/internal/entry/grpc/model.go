package grpc

import (
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
)

func makePbMessage(m *bizp2p.ChatMsg) *pbmsg.Message {
	if m == nil {
		return &pbmsg.Message{}
	}

	return &pbmsg.Message{
		MsgId:    m.MsgId,
		ChatId:   m.ChatId,
		Sender:   m.Sender,
		Receiver: m.Receiver,
		Status:   m.Status,
		Content: &pbmsg.MsgContent{
			Type: m.Type,
			Data: m.Content,
		},
		Seq: m.Seq,
	}
}
