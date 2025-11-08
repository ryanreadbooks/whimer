package sysnotify

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
)

type lazyCheckedMsg interface {
	getMsgId() string
	getNoteId() int64
	getCommentIds() []int64
	shouldRuleOut(noteExistence, commentExistence map[int64]bool) bool
}

type mentionedMsgLazyAdapter struct {
	*model.MentionedMsg
}

func (a *mentionedMsgLazyAdapter) getMsgId() string {
	return a.MentionedMsg.Id
}

func (a *mentionedMsgLazyAdapter) getNoteId() int64 {
	return int64(a.MentionedMsg.NoteId)
}

func (a *mentionedMsgLazyAdapter) getCommentIds() []int64 {
	switch a.MentionedMsg.Type {
	case model.NotifyMsgOnComment:
		return []int64{a.MentionedMsg.CommentId}
	}

	return []int64{}
}

func (a *mentionedMsgLazyAdapter) shouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	noteId := int64(a.MentionedMsg.NoteId)
	switch a.MentionedMsg.Type {
	case model.NotifyMsgOnComment:
		noteOk, _ := noteExistence[noteId]
		commentOk, _ := commentExistence[a.CommentId]
		if !noteOk || !commentOk {
			a.MentionedMsg.DoNotReveal()
			return true
		}
	case model.NotifyMsgOnNote:
		noteOk, _ := noteExistence[noteId]
		if !noteOk {
			a.MentionedMsg.DoNotReveal()
			return true
		}
	}

	return false
}

func getLazyCheckedMsgForMentionedMsgs(msgs []*model.MentionedMsg) []lazyCheckedMsg {
	result := make([]lazyCheckedMsg, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, &mentionedMsgLazyAdapter{m})
	}

	return result
}

type replyMsgLazyAdapter struct {
	*model.ReplyMsg
}

func (a *replyMsgLazyAdapter) getMsgId() string {
	return a.ReplyMsg.Id
}

func (a *replyMsgLazyAdapter) getNoteId() int64 {
	if a.ReplyMsg.Type == model.NotifyMsgOnNote {
		return int64(a.ReplyMsg.NoteId)
	}

	return 0
}

func (a *replyMsgLazyAdapter) getCommentIds() []int64 {
	switch a.ReplyMsg.Type {
	case model.NotifyMsgOnComment:
		return []int64{a.ReplyMsg.TargetComment, a.ReplyMsg.TriggerComment}
	}

	return []int64{}
}

func (a *replyMsgLazyAdapter) shouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	noteId := int64(a.ReplyMsg.NoteId)
	switch a.ReplyMsg.Type {
	case model.NotifyMsgOnComment:
		noteOk, _ := noteExistence[noteId]
		commentOk, _ := commentExistence[a.ReplyMsg.TriggerComment]
		commentOk2 := commentExistence[a.ReplyMsg.TargetComment]
		if !noteOk || !commentOk || !commentOk2 {
			a.ReplyMsg.DoNotReveal()
			return true
		}
	case model.NotifyMsgOnNote:
		noteOk, _ := noteExistence[noteId]
		if !noteOk {
			a.ReplyMsg.DoNotReveal()
			return true
		}
	}

	return false
}

func getLazyCheckedMsgForReplyMsgs(msgs []*model.ReplyMsg) []lazyCheckedMsg {
	result := make([]lazyCheckedMsg, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, &replyMsgLazyAdapter{m})
	}

	return result
}

type likesMsgLazyAdapter struct {
	*model.LikesMsg
}

func (a *likesMsgLazyAdapter) getMsgId() string {
	return a.LikesMsg.Id
}

func (a *likesMsgLazyAdapter) getNoteId() int64 {
	return int64(a.LikesMsg.NoteId)
}

func (a *likesMsgLazyAdapter) getCommentIds() []int64 {
	switch a.LikesMsg.Type {
	case model.NotifyMsgOnComment:
		return []int64{a.LikesMsg.CommentId}
	}

	return []int64{}
}

func (a *likesMsgLazyAdapter) shouldRuleOut(noteExistence, commentExistence map[int64]bool) bool {
	noteId := int64(a.LikesMsg.NoteId)
	switch a.LikesMsg.Type {
	case model.NotifyMsgOnComment:
		noteOk, _ := noteExistence[noteId]
		commentOk, _ := commentExistence[a.LikesMsg.CommentId]
		if !noteOk || !commentOk {
			a.LikesMsg.DoNotReveal()
			return true
		}
	case model.NotifyMsgOnNote:
		noteOk, _ := noteExistence[noteId]
		if !noteOk {
			a.LikesMsg.DoNotReveal()
			return true
		}
	}

	return false
}

func getLazyCheckedMsgForLikesMsgs(msgs []*model.LikesMsg) []lazyCheckedMsg {
	result := make([]lazyCheckedMsg, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, &likesMsgLazyAdapter{m})
	}

	return result
}
