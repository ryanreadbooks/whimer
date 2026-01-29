package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/vo"
)

type CommentAdapter interface {
	// 检查是否评论过
	BatchCheckCommented(ctx context.Context, p *BatchCheckCommentedParams) (*BatchCheckCommentedResult, error)

	// 批量检查评论是否存在
	BatchCheckCommentExist(ctx context.Context, commentIds []int64) (map[int64]bool, error)

	// 批量检查多个用户对多条评论的点赞状态 mappings: uid -> commentIds
	BatchCheckUsersLikeComment(ctx context.Context, mappings map[int64][]int64) (map[int64]map[int64]bool, error)

	// 发布评论
	AddComment(ctx context.Context, p *AddCommentParams) (int64, error)
	// 获取评论
	GetComment(ctx context.Context, commentId int64) (*entity.Comment, error)
	// 获取评论发布者
	GetCommentUser(ctx context.Context, commentId int64) (int64, error)
	// 删除评论
	DelComment(ctx context.Context, commentId, oid int64) error
	// 置顶/取消置顶评论
	PinComment(ctx context.Context, oid, commentId int64, action vo.PinAction) error
	// 点赞评论
	LikeComment(ctx context.Context, commentId int64, action vo.ThumbAction) error
	// 点踩评论
	DislikeComment(ctx context.Context, commentId int64, action vo.ThumbAction) error
	// 获取评论点赞数
	GetCommentLikeCount(ctx context.Context, commentId int64) (int64, error)
	// 批量检查用户是否点赞评论
	BatchCheckUserLikeComment(ctx context.Context, uid int64, commentIds []int64) (map[int64]bool, error)

	// 分页获取主评论
	PageGetRootComments(ctx context.Context, p *PageGetCommentsParams) (*PageGetCommentsResult, error)
	// 分页获取子评论
	PageGetSubComments(ctx context.Context, p *PageGetSubCommentsParams) (*PageGetCommentsResult, error)
	// 分页获取带子评论的主评论
	PageGetDetailedComments(ctx context.Context, p *PageGetDetailedCommentsParams) (*PageGetDetailedCommentsResult, error)
	// 获取置顶评论
	GetPinnedComment(ctx context.Context, oid int64) (*entity.DetailedComment, error)
}

type BatchCheckCommentedParams struct {
	Uid     int64
	NoteIds []int64
}

type BatchCheckCommentedResult struct {
	// noteId -> commented
	Commented map[int64]bool
}

type AddCommentParams struct {
	Type     int32
	Oid      int64
	Content  string
	RootId   int64
	ParentId int64
	ReplyUid int64
	Images   []vo.CommentImage
	AtUsers  []vo.AtUser
}

type PageGetCommentsParams struct {
	Oid    int64
	Cursor int64
	SortBy vo.SortType
}

type PageGetCommentsResult struct {
	Items      []*entity.Comment
	NextCursor int64
	HasNext    bool
}

type PageGetSubCommentsParams struct {
	Oid    int64
	RootId int64
	Cursor int64
}

type PageGetDetailedCommentsParams struct {
	Oid    int64
	Cursor int64
	SortBy vo.SortType
}

type PageGetDetailedCommentsResult struct {
	Items      []*entity.DetailedComment
	NextCursor int64
	HasNext    bool
}
