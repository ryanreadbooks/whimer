package entity

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

// 系统消息 收到的赞
//
// 用于对接各端的收到的赞系统消息结构
type LikesMsg struct {
	Id        string                     `json:"id,omitempty"`      // 消息uuidv7
	SendAt    int64                      `json:"send_at,omitempty"` // 发送时间
	Type      notifyvo.NotifyMsgLocation `json:"type,omitempty"`
	Uid       int64                      `json:"uid,omitempty"`      // 谁点赞
	RecvUid   int64                      `json:"recv_uid,omitempty"` // 谁被点赞
	NoteId    noteid.NoteId              `json:"note_id,omitempty"`
	CommentId int64                      `json:"comment_id,omitempty"`
	Status    notifyvo.MsgStatus         `json:"status"`
}

func (m *LikesMsg) DoNotReveal() {
	if m == nil {
		return
	}

	m.Status = notifyvo.MsgStatusNoReveal
	// m.Id = "" // Id需要保留外部删除使用
	m.SendAt = 0
	m.Type = ""
	m.Uid = 0
	m.NoteId = 0
	m.CommentId = 0
	m.RecvUid = 0
}

func (m *LikesMsg) GetSourceNoteId() int64 {
	return int64(m.NoteId)
}

func (m *LikesMsg) GetSourceCommentIds() []int64 {
	if m.Type == notifyvo.NotifyMsgOnComment {
		return []int64{m.CommentId}
	}
	return nil
}

// ShouldRuleOut 判断消息源是否不存在，需要被过滤掉
func (m *LikesMsg) ShouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	if m == nil {
		return false
	}

	noteId := int64(m.NoteId)
	switch m.Type {
	case notifyvo.NotifyMsgOnComment:
		noteOk := noteExistence[noteId]
		commentOk := commentExistence[m.CommentId]
		if !noteOk || !commentOk {
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

// ShouldFilterByLikeStatus 判断点赞是否已取消，需要被过滤掉
func (m *LikesMsg) ShouldFilterByLikeStatus(noteLikeStatus, commentLikeStatus map[int64]map[int64]bool) bool {
	if m == nil {
		return false
	}

	uid := m.Uid
	switch m.Type {
	case notifyvo.NotifyMsgOnNote:
		noteId := int64(m.NoteId)
		if uidStatus, ok := noteLikeStatus[uid]; ok {
			if liked, exists := uidStatus[noteId]; !exists || !liked {
				m.DoNotReveal()
				return true
			}
		}
	case notifyvo.NotifyMsgOnComment:
		commentId := m.CommentId
		if uidStatus, ok := commentLikeStatus[uid]; ok {
			if liked, exists := uidStatus[commentId]; !exists || !liked {
				m.DoNotReveal()
				return true
			}
		}
	}
	return false
}
