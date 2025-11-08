package model

import (
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

// 系统消息 被回复
//
// 用于对接各端的被回复的系统消息结构
type ReplyMsg struct {
	Id             string            `json:"id,omitempty"` // 消息uuidv7
	SendAt         int64             `json:"send_at,omitempty"`
	Type           NotifyMsgLocation `json:"type,omitempty"`
	Uid            int64             `json:"uid,omitempty"`             // 谁回复的
	RecvUid        int64             `json:"recv_uid,omitempty"`        // 被回复的
	NoteId         imodel.NoteId     `json:"note_id,omitempty"`         // 回复所属笔记
	TargetComment  int64             `json:"target_comment,omitempty"`  // 被回复的评论
	TriggerComment int64             `json:"trigger_comment,omitempty"` // 回复评论的评论
	Content        string            `json:"content,omitempty"`
	Status         MsgStatus         `json:"status"`

	Ext *ReplyMsgExt `json:"ext,omitempty"`
}

type ReplyMsgExt struct {
	AtUsers []imodel.AtUser `json:"at_users,omitempty"`
}

func (m *ReplyMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Content = ""
	m.Status = MsgStatusNoReveal
	m.Id = ""
	m.Type = ""
	m.SendAt = 0
	m.Uid = 0
	m.NoteId = 0
	m.TargetComment = 0
	m.TriggerComment = 0
	m.RecvUid = 0
	m.Ext = nil
}
