package vo

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
)

type NotifyUserReplyParam struct {
	Loc            NotifyMsgLocation `json:"loc"`
	TargetComment  int64             `json:"target,omitempty"` // 被回复的评论
	TriggerComment int64             `json:"trigger"`          // 用这条评论回复的
	SrcUid         int64             `json:"src_uid"`
	RecvUid        int64             `json:"recv_uid"`
	NoteId         noteid.NoteId     `json:"note_id"`
	Content        []byte            `json:"content"` // see CommentContent
}

type CommentContent struct {
	Text    string              `json:"text"`
	AtUsers []*mentionvo.AtUser `json:"at_users"`
}
