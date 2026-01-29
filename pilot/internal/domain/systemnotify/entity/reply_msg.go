package entity

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

// 系统消息 被回复
//
// 用于对接各端的被回复的系统消息结构
type ReplyMsg struct {
	Id             string                     `json:"id,omitempty"` // 消息uuidv7
	SendAt         int64                      `json:"send_at,omitempty"`
	Type           notifyvo.NotifyMsgLocation `json:"type,omitempty"`
	Uid            int64                      `json:"uid,omitempty"`             // 谁回复的
	RecvUid        int64                      `json:"recv_uid,omitempty"`        // 被回复的
	NoteId         noteid.NoteId              `json:"note_id,omitempty"`         // 回复所属笔记
	TargetComment  int64                      `json:"target_comment,omitempty"`  // 被回复的评论
	TriggerComment int64                      `json:"trigger_comment,omitempty"` // 回复评论的评论
	Content        string                     `json:"content,omitempty"`
	Status         notifyvo.MsgStatus         `json:"status"`

	Ext *ReplyMsgExt `json:"ext,omitempty"`
}

type ReplyMsgExt struct {
	AtUsers []*mentionvo.AtUser `json:"at_users,omitempty"`
}

func (m *ReplyMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Content = ""
	m.Status = notifyvo.MsgStatusNoReveal
	// m.Id = "" // id需要保留在外部删除使用
	m.Type = ""
	m.SendAt = 0
	m.Uid = 0
	m.NoteId = 0
	m.TargetComment = 0
	m.TriggerComment = 0
	m.RecvUid = 0
	m.Ext = nil
}

func (m *ReplyMsg) GetSourceNoteId() int64 {
	if m.Type == notifyvo.NotifyMsgOnNote {
		return int64(m.NoteId)
	}
	return 0
}

func (m *ReplyMsg) GetSourceCommentIds() []int64 {
	if m.Type == notifyvo.NotifyMsgOnComment {
		return []int64{m.TargetComment, m.TriggerComment}
	}
	return nil
}

// ShouldRuleOut 判断消息源是否不存在，需要被过滤掉
func (m *ReplyMsg) ShouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	if m == nil {
		return false
	}

	noteId := int64(m.NoteId)
	switch m.Type {
	case notifyvo.NotifyMsgOnComment:
		noteOk := noteExistence[noteId]
		commentOk := commentExistence[m.TriggerComment]
		commentOk2 := commentExistence[m.TargetComment]
		if !noteOk || !commentOk || !commentOk2 {
			m.DoNotReveal()
			return true
		}
	case notifyvo.NotifyMsgOnNote:
		noteOk := noteExistence[noteId]
		if !noteOk {
			m.DoNotReveal()
			return true
		}
	}
	return false
}
