package model

import (
	"unicode/utf8"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
)

// 发表评论参数
type AddCommentReq struct {
	Type     CommentType                  `json:"type"`    // 评论类型 (0-文本; 1-图文)
	Oid      int64                        `json:"nid"`     // 对象id
	Content  string                       `json:"content"` // 评论内容
	RootId   int64                        `json:"pid"`     // 根评论id
	ParentId int64                        `json:"rid"`     // 被回复的评论id
	ReplyUid int64                        `json:"ruid"`    // 被回复的用户id
	Images   []*commentv1.CommentReqImage `json:"images"`
	AtUsers  []*commentv1.CommentAtUser   `json:"at_users"` // @用户列表
}

func (r *AddCommentReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	// 评论类型校验
	if r.Type != CommentText && r.Type != CommentImageText {
		return global.ErrUnsupportedType
	}

	// 评论对象id不能为空
	if r.Oid == 0 {
		return global.ErrObjectIdEmpty
	}

	if r.ReplyUid == 0 {
		return global.ErrReplyUidEmpty
	}

	// 评论长度不能太长或者太短
	cLen := utf8.RuneCountInString(r.Content)
	switch r.Type {
	case CommentText:
		if cLen < MinContentLen {
			return global.ErrContentTooShort
		}
		if cLen > MaxContentLen {
			return global.ErrContentTooLong
		}
	case CommentImageText:
		imageLen := len(r.Images)
		if imageLen <= 0 || imageLen > MaxCommentImageCount {
			return global.ErrInvalidImageCount
		}
	}

	// 评论的关系
	// 子评论一定要附着在主评论下
	if r.ParentId != 0 && r.RootId == 0 {
		return global.ErrCommentWrongRelation
	}

	if len(r.AtUsers) > 0 {
		r.AtUsers = FilterInvalidAtUsers(r.AtUsers)
	}

	return nil
}

// 发表评论结果
type AddCommentRes struct {
	CommentId int64
	Uid       int64
}

func FilterInvalidAtUsers(atUsers []*commentv1.CommentAtUser) []*commentv1.CommentAtUser {
	if len(atUsers) == 0 {
		return nil
	}

	var r []*commentv1.CommentAtUser
	for _, a := range atUsers {
		if len(a.Nickname) == 0 || a.Uid == 0 {
			continue
		}

		r = append(r, a)
	}

	return r
}
