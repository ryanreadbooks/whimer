package dto

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/errors"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// GetCommentsQuery 获取评论查询
type GetCommentsQuery struct {
	Oid    notevo.NoteId `form:"oid"`
	Cursor int64         `form:"cursor,optional"`
	SortBy int           `form:"sort_by,optional"`
	SeekId int64         `form:"seek_id,optional"`
}

func (q *GetCommentsQuery) Validate() error {
	if q.Oid <= 0 {
		return errors.ErrNoteNotFound
	}
	return nil
}

// GetSubCommentsQuery 获取子评论查询
type GetSubCommentsQuery struct {
	Oid    notevo.NoteId `form:"oid"`
	RootId int64         `form:"root"`
	Cursor int64         `form:"cursor,optional"`
}

func (q *GetSubCommentsQuery) Validate() error {
	if q.Oid <= 0 {
		return errors.ErrNoteNotFound
	}
	if q.RootId <= 0 {
		return errors.ErrCommentNotFound
	}
	return nil
}

// GetLikeCountQuery 获取点赞数查询
type GetLikeCountQuery struct {
	CommentId int64 `form:"comment_id"`
}

func (q *GetLikeCountQuery) Validate() error {
	if q.CommentId <= 0 {
		return errors.ErrCommentNotFound
	}
	return nil
}

// GetLikeCountResult 获取点赞数结果
type GetLikeCountResult struct {
	CommentId int64 `json:"comment_id"`
	Likes     int64 `json:"likes"`
}
