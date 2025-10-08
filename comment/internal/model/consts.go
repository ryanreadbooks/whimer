package model

import commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"

// 评论类型
type CommentType int8

const (
	CommentText        = CommentType(commentv1.CommentType_Text)        // 纯文本
	CommentImageText   = CommentType(commentv1.CommentType_ImageText)   // 图文
	CommentCustomEmoji = CommentType(commentv1.CommentType_CustomEmoji) // 自定义表情
)

// 评论是否置顶
const (
	IsNotPinned = 0
	IsPinned    = 1
)

// 评论状态
type CommentState int8

// 评论状态
const (
	// TODO define more comment state
	CommentStateNormal CommentState = 0
)

// 评论资源类型
type CommentAssetType int8

const (
	CommentAssetImage       = 1
	CommentAssetCustomEmoji = 2
)

func (t CommentAssetType) String() string {
	switch t {
	case CommentAssetImage:
		return "inline_image"
	case CommentAssetCustomEmoji:
		return "custom_emoji"
	default:
		return "unknown"
	}
}

const (
	minContentLen = 1
	maxContentLen = 2000

	MaxCommentImageCount = 9
)
