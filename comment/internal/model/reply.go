package model

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
)

// 评论类型
type ReplyType int8

const (
	ReplyText      ReplyType = 0
	ReplyImageText ReplyType = 1
)

// 评论是否置顶
const (
	IsNotPinned = 0
	IsPinned    = 1
)

// 评论状态
type ReplyState int8

// 评论状态
const (
	// TODO define more reply state
	ReplyStateNormal ReplyState = 0
)

const (
	minContentLen = 1
	maxContentLen = 2000
)

// 发表评论参数
type ReplyReq struct {
	Type     ReplyType `json:"type"`    // 评论类型 (0-文本; 1-图文)
	Oid      uint64    `json:"nid"`     // 对象id
	Content  string    `json:"content"` // 评论内容
	RootId   uint64    `json:"pid"`     // 根评论id
	ParentId uint64    `json:"rid"`     // 被回复的评论id
	ReplyUid uint64    `json:"ruid"`    // 被回复的用户id
}

func (r *ReplyReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.Type != ReplyText && r.Type != ReplyImageText {
		return global.ErrUnsupportedType
	}

	if r.Oid <= 0 {
		return global.ErrObjectIdEmpty
	}

	cLen := utf8.RuneCountInString(r.Content)
	if r.Type == ReplyText {
		if cLen < minContentLen {
			return global.ErrContentTooShort
		}
		if cLen > maxContentLen {
			return global.ErrContentTooLong
		}
	}

	return nil
}

// 发表评论结果
type ReplyRes struct {
	ReplyId uint64
	Uid     uint64
}
