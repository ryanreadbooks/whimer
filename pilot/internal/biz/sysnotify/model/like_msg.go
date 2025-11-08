package model

import (
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

// 系统消息 收到的赞
//
// 用于对接各端的收到的赞系统消息结构
type LikesMsg struct {
	Id        string            `json:"id,omitempty"`      // 消息uuidv7
	SendAt    int64             `json:"send_at,omitempty"` // 发送时间
	Type      NotifyMsgLocation `json:"type,omitempty"`
	Uid       int64             `json:"uid,omitempty"`      // 谁点赞
	RecvUid   int64             `json:"recv_uid,omitempty"` // 谁被点赞
	NoteId    imodel.NoteId     `json:"note_id,omitempty"`
	CommentId int64             `json:"comment_id,omitempty"`
	Status    MsgStatus         `json:"status"`
}

func (m *LikesMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Status = MsgStatusNoReveal
	m.Id = ""
	m.SendAt = 0
	m.Type = ""
	m.Uid = 0
	m.NoteId = 0
	m.CommentId = 0
	m.RecvUid = 0
}
