package model

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
)

type ReplyType int8

const (
	ReplyText      ReplyType = 0
	ReplyImageText ReplyType = 1
)

const (
	minContentLen = 1
	maxContentLen = 2000
)

// 发表评论参数
type ReplyReq struct {
	Type     ReplyType `json:"type" form:"type"`         // 评论类型 (0-文本; 1-图文)
	NoteId   string    `json:"nid" form:"nid"`           // 笔记id
	Content  string    `json:"content" form:"content"`   // 评论内容
	ParentId uint64    `json:"pid,omitempty" form:"pid"` // 根评论id
	ReplyId  uint64    `json:"rid,omitempty" form:"rid"` // 被回复的评论id
}

func (r *ReplyReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.Type != ReplyText && r.Type != ReplyImageText {
		return global.ErrUnsupportedType
	}

	if len(r.NoteId) == 0 {
		return global.ErrNoteIdEmpty
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
