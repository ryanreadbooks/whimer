package system

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	systemdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/system"
)

type CreateSystemMsgReq struct {
	MsgId      uuid.UUID // 请求时无需设置
	TriggerUid int64     // 触发系统消息的用户uid
	RecvUid    int64     // 接收者uid
	ChatType   model.SystemChatType
	MsgType    model.MsgType
	Content    []byte
}

type ListMsgReq struct {
	RecvUid  int64
	ChatType model.SystemChatType
	Cursor   string // msg主键uuid的string形式
	Count    int32
}

type SystemMsg struct {
	Id           uuid.UUID
	SystemChatId uuid.UUID
	TriggerUid   int64
	RecvUid      int64
	Status       model.SystemMsgStatus
	MsgType      model.MsgType
	Content      []byte
	Mtime        int64
}

type SystemChat struct {
	Id            uuid.UUID
	Type          model.SystemChatType
	Uid           int64
	Mtime         int64
	LastMsgId     uuid.UUID
	LastReadMsgId uuid.UUID
	LastReadTime  int64
	UnreadCount   int64
	LastMsg       *SystemMsg // 最后一条消息
}

func MakeSystemMsgFromPO(po *systemdao.MsgPO) *SystemMsg {
	content := po.Content
	if po.Status == model.SystemMsgStatusRevoked {
		content = []byte("消息已被撤回")
	}
	return &SystemMsg{
		Id:           po.Id,
		SystemChatId: po.SystemChatId,
		TriggerUid:   po.Uid,
		RecvUid:      po.RecvUid,
		Status:       po.Status,
		MsgType:      po.MsgType,
		Content:      content,
		Mtime:        po.Mtime,
	}
}

func MakeSystemMsgsFromPOs(pos []*systemdao.MsgPO) []*SystemMsg {
	if len(pos) == 0 {
		return []*SystemMsg{}
	}

	msgs := make([]*SystemMsg, 0, len(pos))
	for _, po := range pos {
		msgs = append(msgs, MakeSystemMsgFromPO(po))
	}
	return msgs
}

func MakeSystemChatFromPO(po *systemdao.ChatPO) *SystemChat {
	return &SystemChat{
		Id:            po.Id,
		Type:          po.Type,
		Uid:           po.Uid,
		Mtime:         po.Mtime,
		LastMsgId:     po.LastMsgId,
		LastReadMsgId: po.LastReadMsgId,
		LastReadTime:  po.LastReadTime,
		UnreadCount:   po.UnreadCount,
	}
}

type ChatUnread struct {
	ChatId      uuid.UUID
	ChatType    model.SystemChatType
	UnreadCount int64
}

func ChatUnreadFromPo(c *systemdao.ChatPO) *ChatUnread {
	return &ChatUnread{
		ChatId:      c.Id,
		ChatType:    c.Type,
		UnreadCount: c.UnreadCount,
	}
}
