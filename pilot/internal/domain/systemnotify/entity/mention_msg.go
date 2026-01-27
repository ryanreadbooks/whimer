package entity

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	vo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

// 系统消息 被@
//
// 用于对接各端的被@的系统消息结构
type MentionedMsg struct {
	Id        string               `json:"id,omitempty"`      // 消息uuidv7
	SendAt    int64                `json:"send_at,omitempty"` // 发送时间
	Type      vo.NotifyMsgLocation `json:"type,omitempty"`
	Uid       int64                `json:"uid,omitempty"`        // 谁@的
	RecvUsers []*mentionvo.AtUser  `json:"recv_users,omitempty"` // 被@的
	NoteId    noteid.NoteId        `json:"note_id,omitempty"`    // 从哪篇笔记@的
	CommentId int64                `json:"comment_id,omitempty"` // 从哪条评论@的
	Content   string               `json:"content,omitempty"`    // 内容 笔记中的desc或者comment
	Status    vo.MsgStatus         `json:"status"`
}

func (m *MentionedMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Content = ""
	m.Status = vo.MsgStatusNoReveal
	// m.Id = "" // id需要保留在外部删除使用
	m.SendAt = 0
	m.Type = ""
	m.Uid = 0
	m.NoteId = 0
	m.CommentId = 0
	m.RecvUsers = nil
}

func (m *MentionedMsg) GetSourceNoteId() int64 {
	return int64(m.NoteId)
}

func (m *MentionedMsg) GetSourceCommentIds() []int64 {
	if m.Type == vo.NotifyMsgOnComment {
		return []int64{m.CommentId}
	}
	return nil
}

// ShouldRuleOut 判断消息源是否不存在，需要被过滤掉
func (m *MentionedMsg) ShouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	if m == nil {
		return false
	}

	noteId := int64(m.NoteId)
	switch m.Type {
	case vo.NotifyMsgOnComment: // 评论中@
		noteOk := noteExistence[noteId]
		commentOk := commentExistence[m.CommentId]
		if !noteOk || !commentOk {
			m.DoNotReveal()
			return true
		}
	case vo.NotifyMsgOnNote: // 笔记中@
		noteOk := noteExistence[noteId]
		if !noteOk {
			m.DoNotReveal()
			return true
		}
	}
	return false
}

type ChatUnread struct {
	ChatId string `json:"chat_id"`
	Count  int64  `json:"count"`
}

type ChatsUnreadCount struct {
	MentionUnread ChatUnread `json:"mention_unread"`
	NoticeUnread  ChatUnread `json:"notice_unread"`
	LikesUnread   ChatUnread `json:"likes_unread"`
	ReplyUnread   ChatUnread `json:"reply_unread"`
}
