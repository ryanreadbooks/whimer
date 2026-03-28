package dto

import (
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

// MsgType 消息类型
type MsgType string

const (
	MsgTypeText  MsgType = "text"
	MsgTypeImage MsgType = "image"
)

// IsValid 检查消息类型是否有效
func (t MsgType) IsValid() bool {
	return t == MsgTypeText || t == MsgTypeImage
}

// ToVO 转换为 vo.MsgType
func (t MsgType) ToVO() vo.MsgType {
	switch t {
	case MsgTypeText:
		return vo.MsgText
	case MsgTypeImage:
		return vo.MsgImage
	default:
		return vo.MsgTypeUnspecified
	}
}

// MsgTypeFromVO 从 vo.MsgType 转换为 MsgType
func MsgTypeFromVO(t vo.MsgType) MsgType {
	switch t {
	case vo.MsgText:
		return MsgTypeText
	case vo.MsgImage:
		return MsgTypeImage
	default:
		return ""
	}
}

// MsgStatus 消息状态
type MsgStatus string

const (
	MsgStatusNormal   MsgStatus = "normal"
	MsgStatusRecalled MsgStatus = "recalled"
)

// MsgStatusFromVO 从 vo.MsgStatus 转换为 MsgStatus
func MsgStatusFromVO(s vo.MsgStatus) MsgStatus {
	switch s {
	case vo.MsgStatusNormal:
		return MsgStatusNormal
	case vo.MsgStatusRecall:
		return MsgStatusRecalled
	default:
		return ""
	}
}

type MsgWithSender struct {
	Id        string         `json:"id,omitempty"`
	Cid       string         `json:"cid,omitempty"`
	Type      MsgType        `json:"type,omitempty"`
	Status    MsgStatus      `json:"status,omitempty"`
	Mtime     int64          `json:"mtime,omitempty"`
	SenderUid int64          `json:"sender_uid,omitempty"`
	Content   *MsgContentDto `json:"content,omitempty"`
	Pos       int64          `json:"pos"`
	Ext       *MsgExtDto     `json:"ext,omitempty"`
	Sender    *uservo.User   `json:"sender,omitempty"`
}

type MsgContentDto struct {
	Text  *MsgTextContentDto  `json:"text,omitempty"`
	Image *MsgImageContentDto `json:"image,omitempty"`
}

type MsgTextContentDto struct {
	Content string `json:"content"`
	Preview string `json:"preview,omitempty"`
}

type MsgImageContentDto struct {
	Key     string `json:"key"`
	Height  uint32 `json:"height"`
	Width   uint32 `json:"width"`
	Format  string `json:"format"`
	Preview string `json:"preview,omitempty"`
}

type MsgExtDto struct {
	Recall *MsgExtRecallDto `json:"recall,omitempty"`
}

type MsgExtRecallDto struct {
	RecallUid int64 `json:"recall_uid"`
	RecallAt  int64 `json:"recall_at"`
}

func MsgWithSenderFromEntity(msg *entity.Msg) *MsgWithSender {
	if msg == nil {
		return nil
	}
	result := &MsgWithSender{
		Id:        msg.Id,
		Cid:       msg.Cid,
		Type:      MsgTypeFromVO(msg.Type),
		Status:    MsgStatusFromVO(msg.Status),
		Mtime:     msg.Mtime,
		SenderUid: msg.SenderUid,
		Pos:       msg.Pos,
	}
	if msg.Content != nil {
		result.Content = &MsgContentDto{}
		if msg.Content.Text != nil {
			result.Content.Text = &MsgTextContentDto{
				Content: msg.Content.Text.Content,
				Preview: msg.Content.Text.Preview,
			}
		}
		if msg.Content.Image != nil {
			result.Content.Image = &MsgImageContentDto{
				Key:     msg.Content.Image.Key,
				Height:  msg.Content.Image.Height,
				Width:   msg.Content.Image.Width,
				Format:  msg.Content.Image.Format,
				Preview: msg.Content.Image.Preview,
			}
		}
	}
	if msg.Ext != nil {
		result.Ext = &MsgExtDto{}
		if msg.Ext.Recall != nil {
			result.Ext.Recall = &MsgExtRecallDto{
				RecallUid: msg.Ext.Recall.RecallUid,
				RecallAt:  msg.Ext.Recall.RecallAt,
			}
		}
	}
	return result
}

type RecentChatWithPeer struct {
	ChatId      string          `json:"chat_id"`
	ChatType    vo.ChatType     `json:"chat_type"`
	ChatName    string          `json:"chat_name,omitempty"`
	ChatCreator int64           `json:"chat_creator,omitempty"`
	LastMsg     *MsgWithSender  `json:"last_msg,omitempty"`
	UnreadCount int64           `json:"unread_count"`
	Mtime       int64           `json:"mtime"`
	IsPinned    bool            `json:"is_pinned"`
	Cover       string          `json:"cover"`
	Peer        *commondto.User `json:"peer,omitempty"`
}

type ListRecentChatsResult struct {
	Chats      []*RecentChatWithPeer `json:"items"`
	NextCursor string                `json:"next_cursor"`
	HasNext    bool                  `json:"has_next"`
}
