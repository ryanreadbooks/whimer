package repository

import "context"

type CheckCommentedParams struct {
	Uid     int64
	NoteIds []int64
}

type CheckCommentedResult struct {
	// noteId -> commented
	Commented map[int64]bool
}

type CommentAdapter interface {
	// 检查是否评论过
	CheckCommented(ctx context.Context, p *CheckCommentedParams) (*CheckCommentedResult, error)
}
