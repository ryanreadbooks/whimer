package p2p

import (
	p2pdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatMsg struct {
	MsgId    int64
	Sender   int64
	Receiver int64
	ChatId   int64
	Type     model.MsgType
	Status   model.MsgStatus
	Content  string
	Seq      int64
}

func (m *ChatMsg) IsRevoked() bool {
	return m.Status == model.MsgStatusRevoked
}

func MakeChatMsgFromPO(po *p2pdao.MsgPO, recv int64) *ChatMsg {
	if po == nil {
		return &ChatMsg{}
	}

	// po中没有记录receiver
	cm := &ChatMsg{
		MsgId:    po.MsgId,
		Sender:   po.SenderId,
		Receiver: recv,
		ChatId:   po.ChatId,
		Type:     model.MsgType(po.MsgType),
		Status:   model.MsgStatus(po.Status),
		Seq:      po.Seq,
		Content:  po.Content,
	}

	if cm.Status == model.MsgStatusRevoked {
		cm.Content = "" // 已撤回
		cm.Type = model.MsgTypeUnknown
	}

	return cm
}

type CreateMsgReq struct {
	ChatId   int64
	Sender   int64
	Receiver int64
	MsgType  model.MsgType
	Content  string
}

// 单聊会话
type Chat struct {
	ChatId        int64
	UserId        int64
	PeerId        int64 // 需要额外赋值
	Unread        int64
	LastMsgId     int64
	LastMsgSeq    int64
	LastReadMsgId int64
	LastReadTime  int64
	LastMsg       *ChatMsg // 需要额外赋值
}

func MakeChatFromPO(po *p2pdao.ChatPO) *Chat {
	return &Chat{
		ChatId:        po.ChatId,
		UserId:        po.UserId,
		PeerId:        po.PeerId,
		Unread:        po.UnreadCount,
		LastMsgId:     po.LastMsgId,
		LastMsgSeq:    po.LastMsgSeq,
		LastReadMsgId: po.LastReadMsgId,
		LastReadTime:  po.LastReadTime,
	}
}
