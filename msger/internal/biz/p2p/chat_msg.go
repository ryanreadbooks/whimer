package p2p

import (
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
	p2pdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/p2p"
)

type ChatMsg struct {
	MsgId    int64
	Sender   int64
	Receiver int64
	ChatId   int64
	Type     gm.MsgType
	Status   gm.MsgStatus
	Content  string
	Seq      int64
}

func MakeChatMsgFromPO(po *p2pdao.MessagePO, recv int64) *ChatMsg {
	if po == nil {
		return &ChatMsg{}
	}

	// po中没有记录receiver
	cm := &ChatMsg{
		MsgId:    po.MsgId,
		Sender:   po.SenderId,
		Receiver: recv,
		ChatId:   po.ChatId,
		Type:     gm.MsgType(po.MsgType),
		Status:   gm.MsgStatus(po.Status),
		Seq:      po.Seq,
		Content:  po.Content,
	}

	if cm.Status == gm.MsgStatusRevoked {
		cm.Content = "" // 已撤回
	}

	return cm
}

type CreateMsgReq struct {
	ChatId   int64
	Sender   int64
	Receiver int64
	MsgType  gm.MsgType
	Content  string
}
