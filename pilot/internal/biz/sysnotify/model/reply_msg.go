package model

import (
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type ReplyLocation string

const (
	ReplyOnNote    ReplyLocation = "on_note"    // 对笔记发表评论
	ReplyOnComment ReplyLocation = "on_comment" // 对评论发表评论（回复评论）
)

// 系统消息 被回复
//
// 用于对接各端的被回复的系统消息结构
type ReplyMsg struct {
	Id             string        `json:"id"` // 消息uuidv7
	SendAt         int64         `json:"send_at"`
	Type           ReplyLocation `json:"type"`
	Uid            int64         `json:"uid"`                       // 谁回复的
	RecvUid        int64         `json:"recv_uid"`                  // 被回复的
	NoteId         imodel.NoteId `json:"note_id"`                   // 回复所属笔记
	TargetComment  int64         `json:"target_comment,omitempty"`  // 被回复的评论
	TriggerComment int64         `json:"trigger_comment,omitempty"` // 回复评论的评论
	Content        string        `json:"content"`
	Status         MsgStatus     `json:"status"`

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
}
