package p2p

import (
	"github.com/ryanreadbooks/whimer/msger/api/msg"
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

func (m *ChatMsg) IsRevoked() bool {
	return m.Status == gm.MsgStatusRevoked
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
		cm.Type = msg.MsgType_MSG_TYPE_UNSPECIFIED
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

// 单聊会话
type Chat struct {
	ChatId        int64
	UserId        int64
	PeerId        int64
	Unread        int64
	LastMsgId     int64
	LastMsgSeq    int64
	LastReadMsgId int64
	LastReadTime  int64
}

func MakeChatFromPO(po *p2pdao.ChatPO) *Chat {
	return &Chat{
		ChatId:        po.ChatId,
		UserId:        po.UserId,
		PeerId:        po.PeerId,
		Unread:        po.UnreadCount,
		LastMsgId:     po.LastMessageId,
		LastMsgSeq:    po.LastMessageSeq,
		LastReadMsgId: po.LastReadMessageId,
		LastReadTime:  po.LastReadTime,
	}
}
