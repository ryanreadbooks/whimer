package systemnotify

import (
	notifyentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
)

// 提取@消息的源ID
func ExtractMentionMsgSourceIds(msgs []*notifyentity.MentionedMsg) (noteIds, commentIds []int64) {
	noteIds = make([]int64, 0, len(msgs))
	commentIds = make([]int64, 0, len(msgs))
	for _, msg := range msgs {
		if noteId := msg.GetSourceNoteId(); noteId != 0 {
			noteIds = append(noteIds, noteId)
		}
		if ids := msg.GetSourceCommentIds(); len(ids) > 0 {
			commentIds = append(commentIds, ids...)
		}
	}
	return noteIds, commentIds
}

// 提取回复消息的源ID
func ExtractReplyMsgSourceIds(msgs []*notifyentity.ReplyMsg) (noteIds, commentIds []int64) {
	noteIds = make([]int64, 0, len(msgs))
	commentIds = make([]int64, 0, len(msgs))
	for _, msg := range msgs {
		if noteId := msg.GetSourceNoteId(); noteId != 0 {
			noteIds = append(noteIds, noteId)
		}
		if ids := msg.GetSourceCommentIds(); len(ids) > 0 {
			commentIds = append(commentIds, ids...)
		}
	}
	return noteIds, commentIds
}

// 提取点赞消息的源ID
func ExtractLikesMsgSourceIds(msgs []*notifyentity.LikesMsg) (noteIds, commentIds []int64) {
	noteIds = make([]int64, 0, len(msgs))
	commentIds = make([]int64, 0, len(msgs))
	for _, msg := range msgs {
		if noteId := msg.GetSourceNoteId(); noteId != 0 {
			noteIds = append(noteIds, noteId)
		}
		if ids := msg.GetSourceCommentIds(); len(ids) > 0 {
			commentIds = append(commentIds, ids...)
		}
	}
	return noteIds, commentIds
}
