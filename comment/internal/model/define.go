package model

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

// 评论类型
type CommentType int8

const (
	CommentText        = CommentType(commentv1.CommentType_TEXT)         // 纯文本
	CommentImageText   = CommentType(commentv1.CommentType_IMAGE_TEXT)   // 图文
	CommentCustomEmoji = CommentType(commentv1.CommentType_CUSTOM_EMOJI) // 自定义表情
)

func CommentTypeFromPb(t commentv1.CommentType) (CommentType, error) {
	switch t {
	case commentv1.CommentType_TEXT:
		return CommentText, nil
	case commentv1.CommentType_IMAGE_TEXT:
		return CommentImageText, nil
	case commentv1.CommentType_CUSTOM_EMOJI:
		return CommentCustomEmoji, nil
	default:
		return 0, xerror.ErrArgs.Msg("unsupported comment type")
	}
}

func CommentTypeToPb(t CommentType) commentv1.CommentType {
	switch t {
	case CommentText:
		return commentv1.CommentType_TEXT
	case CommentImageText:
		return commentv1.CommentType_IMAGE_TEXT
	case CommentCustomEmoji:
		return commentv1.CommentType_CUSTOM_EMOJI
	}

	return commentv1.CommentType_TEXT
}

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
	CommentAssetImage       CommentAssetType = 1
	CommentAssetCustomEmoji CommentAssetType = 2
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
	MinContentLen        = 1
	MaxContentLen        = 2000
	MaxCommentImageCount = 9
)
