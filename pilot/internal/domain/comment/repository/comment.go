package repository

import "context"

type BatchCheckCommentedParams struct {
	Uid     int64
	NoteIds []int64
}

type BatchCheckCommentedResult struct {
	// noteId -> commented
	Commented map[int64]bool
}

type CommentAdapter interface {
	// 检查是否评论过
	BatchCheckCommented(ctx context.Context, p *BatchCheckCommentedParams) (*BatchCheckCommentedResult, error)
}
