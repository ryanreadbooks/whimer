package vo

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"

	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
)

type NotifyAtUsersOnNoteParam struct {
	Uid         int64                            `json:"uid"`
	TargetUsers []*mentionvo.AtUser              `json:"target_users"`
	Content     *NotifyAtUsersOnNoteParamContent `json:"content"`
}

type NotifyAtUsersOnNoteParamContent struct {
	SourceUid int64         `json:"src_uid"` // trigger uid
	NoteDesc  string        `json:"desc"`
	NoteId    noteid.NoteId `json:"id"` // 笔记id
}

type NotifyAtUsersOnCommentParam struct {
	Uid         int64                               `json:"uid"`          // 谁@
	TargetUsers []*mentionvo.AtUser                 `json:"target_users"` // 谁被@
	Content     *NotifyAtUsersOnCommentParamContent `json:"content"`
}

type NotifyAtUsersOnCommentParamContent struct {
	SourceUid int64         `json:"src_uid"`    // 评论发布者uid
	Comment   string        `json:"comment"`    // 评论内容
	NoteId    noteid.NoteId `json:"note_id"`    // 评论归属笔记id
	CommentId int64         `json:"comment_id"` // 评论id
	RootId    int64         `json:"root_id"`    // 根评论id
	ParentId  int64         `json:"parent_id"`  // 父评论id
}

