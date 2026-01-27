package vo

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
)

type NotifyLikesOnNoteParam struct {
	NoteId noteid.NoteId `json:"note_id"`
}

type NotifyLikesOnCommentParam struct {
	NoteId    noteid.NoteId `json:"note_id"`
	CommentId int64         `json:"comment_id"`
}

// LikesMessage 点赞消息内容
type LikesMessage struct {
	*NotifyLikesOnNoteParam    `json:"note_content,omitempty"`
	*NotifyLikesOnCommentParam `json:"comment_content,omitempty"`
	Loc                        NotifyMsgLocation `json:"loc"`
	Uid                        int64             `json:"uid"`      // 谁点赞
	RecvUid                    int64             `json:"recv_uid"` // 谁被点赞
}

// NewLikesOnNoteMessage 创建笔记点赞消息
func NewLikesOnNoteMessage(uid, recvUid int64, param *NotifyLikesOnNoteParam) *LikesMessage {
	return &LikesMessage{
		NotifyLikesOnNoteParam: param,
		Loc:                    NotifyMsgOnNote,
		Uid:                    uid,
		RecvUid:                recvUid,
	}
}

// NewLikesOnCommentMessage 创建评论点赞消息
func NewLikesOnCommentMessage(uid, recvUid int64, param *NotifyLikesOnCommentParam) *LikesMessage {
	return &LikesMessage{
		NotifyLikesOnCommentParam: param,
		Loc:                       NotifyMsgOnComment,
		Uid:                       uid,
		RecvUid:                   recvUid,
	}
}
