package model

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
)

type MentionLocation string

const (
	MentionOnNote    MentionLocation = "on_note"
	MentionOnComment MentionLocation = "on_comment"
)

type MentionedRecvUser struct {
	Uid      int64  `json:"uid"`
	Nickname string `json:"nickname"`
}

// 系统消息 被@
//
// 用于对接各端
type MentionedMsg struct {
	Id        string             `json:"id"`
	SendAt    int64              `json:"send_at"` // 发送时间
	Type      MentionLocation    `json:"type"`
	Uid       int64              `json:"uid"`       // 谁@的
	RecvUser  *MentionedRecvUser `json:"recv_user"` // 被@的
	NoteId    model.NoteId       `json:"note_id,omitempty"`
	CommentId int64              `json:"comment_id,omitempty"`
	Content   string             `json:"content"` // 内容 笔记中的desc或者comment
}

type ChatUnread struct {
	ChatId string `json:"chat_id"`
	Count  int64  `json:"count"`
}

func ChatUnreadFromPb(pb *v1.ChatUnread) ChatUnread {
	return ChatUnread{
		ChatId: pb.GetChatId(),
		Count:  pb.GetUnreadCount(),
	}
}

type ChatsUnreadCount struct {
	MentionUnread ChatUnread `json:"mention_unread"`
	NoticeUnread  ChatUnread `json:"notice_unread"`
	LikesUnread   ChatUnread `json:"likes_unread"`
	ReplyUnread   ChatUnread `json:"reply_unread"`
}
