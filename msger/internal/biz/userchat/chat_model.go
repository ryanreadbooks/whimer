package userchat

import (
	"slices"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

// 会话定义
type Chat struct {
	Id        uuid.UUID
	Type      model.ChatType
	Name      string // 仅在群聊时有效
	Status    model.ChatStatus
	Creator   int64 // 仅在群聊时有效
	Mtime     int64
	LastMsgId uuid.UUID // 最后一条消息ID
	Settings  int64     // 会话设置 单聊不生效

	Members []int64 // 会话成员 需要额外填充
}

func (c *Chat) IsUserInChat(uid int64) bool {
	return c != nil && slices.Contains(c.Members, uid)
}

func (c *Chat) IsStatusNormal() bool {
	return c != nil && c.Status == model.ChatStatusNormal
}

func (c *Chat) IsP2PChat() bool {
	return c != nil && c.Type == model.P2PChat
}

func (c *Chat) IsGroupChat() bool {
	return c != nil && c.Type == model.GroupChat
}

func makeChatFromPO(p *chat.ChatPO) (c *Chat) {
	c = &Chat{
		Id:        p.Id,
		Type:      p.Type,
		Name:      p.Name,
		Status:    p.Status,
		Creator:   p.Creator,
		Mtime:     p.Mtime,
		LastMsgId: p.LastMsgId,
		Settings:  p.Settings,
	}
	return
}
