package model

import (
	v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type MentionedRecvUser struct {
	Uid      int64  `json:"uid"`
	Nickname string `json:"nickname"`
}

// 系统消息 被@
//
// 用于对接各端的被@的系统消息结构
type MentionedMsg struct {
	Id        string               `json:"id,omitempty"`      // 消息uuidv7
	SendAt    int64                `json:"send_at,omitempty"` // 发送时间
	Type      NotifyMsgLocation    `json:"type,omitempty"`
	Uid       int64                `json:"uid,omitempty"`        // 谁@的
	RecvUsers []*MentionedRecvUser `json:"recv_users,omitempty"` // 被@的
	NoteId    imodel.NoteId        `json:"note_id,omitempty"`    // 从哪篇笔记@的
	CommentId int64                `json:"comment_id,omitempty"` // 从哪条评论@的
	Content   string               `json:"content,omitempty"`    // 内容 笔记中的desc或者comment
	Status    MsgStatus            `json:"status"`
}

func (m *MentionedMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Content = ""
	m.Status = MsgStatusNoReveal
	m.Id = ""
	m.SendAt = 0
	m.Type = ""
	m.Uid = 0
	m.NoteId = 0
	m.CommentId = 0
	m.RecvUsers = nil
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
