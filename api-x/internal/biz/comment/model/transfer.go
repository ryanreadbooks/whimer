package model

import (
	"fmt"
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/model/errors"
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	maxContentLen = 2000
)

type PubReq struct {
	CommentType int32        `json:"type"`
	Oid         model.NoteId `json:"oid"`
	Content     string       `json:"content"`
	RootId      int64        `json:"root_id,omitempty,optional"`
	ParentId    int64        `json:"parent_id,omitempty,optional"`
	ReplyUid    int64        `json:"reply_uid"`

	// optional
	Images []*CommentImage `json:"images,omitempty,optional"`
}

func (r *PubReq) Validate() error {
	if r.Oid == 0 {
		return errors.ErrNoteNotFound
	}

	contentLen := utf8.RuneCountInString(r.Content)
	if contentLen > maxContentLen {
		return xerror.ErrArgs.Msg("评论内容太长")
	}

	switch r.CommentType {
	case int32(commentv1.CommentType_Text):
		if contentLen <= 0 {
			return xerror.ErrArgs.Msg("评论内容为空")
		}
	case int32(commentv1.CommentType_ImageText):
		if len(r.Images) == 0 {
			return xerror.ErrArgs.Msg("无评论图片")
		}
		if len(r.Images) > maxNumCommentImages {
			return xerror.ErrArgs.Msg(fmt.Sprintf("最多支持%d张评论图片", maxNumCommentImages))
		}

		for _, img := range r.Images {
			if img.StoreKey == "" {
				return xerror.ErrArgs.Msg("非法storeKey")
			}

			if img.Width == 0 || img.Height == 0 || img.Format == "" {
				return xerror.ErrArgs.Msg("上传图片未指定图片信息")
			}

			if err := model.CheckImageFormat(img.Format); err != nil {
				return err
			}
		}
	default:
		return xerror.ErrArgs.Msg("不支持的评论类型")
	}

	return nil
}

func (r *PubReq) AsPb() *commentv1.AddCommentRequest {
	images := make([]*commentv1.CommentReqImage, 0, len(r.Images))
	for _, img := range r.Images {
		images = append(images, &commentv1.CommentReqImage{
			StoreKey: img.StoreKey,
			Width:    img.Width,
			Height:   img.Height,
			Format:   img.Format,
		})
	}

	return &commentv1.AddCommentRequest{
		Type:     commentv1.CommentType(r.CommentType),
		Oid:      int64(r.Oid),
		Content:  r.Content,
		RootId:   r.RootId,
		ParentId: r.ParentId,
		ReplyUid: r.ReplyUid,
		Images:   images,
	}
}

type PubRes struct {
	CommentId int64 `json:"comment_id"`
}

type GetCommentsReq struct {
	Oid    model.NoteId `form:"oid"`
	Cursor int64        `form:"cursor,optional"`
	SortBy int          `form:"sort_by,optional"`
}

func (r *GetCommentsReq) AsPb() *commentv1.PageGetCommentRequest {
	return &commentv1.PageGetCommentRequest{
		Oid:    int64(r.Oid),
		Cursor: r.Cursor,
		SortBy: commentv1.SortType(r.SortBy),
	}
}

func (r *GetCommentsReq) AsDetailedPb() *commentv1.PageGetDetailedCommentRequest {
	return &commentv1.PageGetDetailedCommentRequest{
		Oid:    int64(r.Oid),
		Cursor: r.Cursor,
		SortBy: commentv1.SortType(r.SortBy),
	}
}

type CommentRes struct {
	Items      []*CommentItem `json:"items"`
	NextCursor int64          `json:"next_cursor"`
	HasNext    bool           `json:"has_next"`
}

type GetSubCommentsReq struct {
	Oid    model.NoteId `form:"oid"`
	RootId int64        `form:"root"`
	Cursor int64        `form:"cursor,optional"`
}

func (r *GetSubCommentsReq) AsPb() *commentv1.PageGetSubCommentRequest {
	return &commentv1.PageGetSubCommentRequest{
		Oid:    int64(r.Oid),
		RootId: r.RootId,
		Cursor: r.Cursor,
	}
}

type DetailedCommentRes struct {
	Comments   []*DetailedCommentItem `json:"comments"`
	PinComment *DetailedCommentItem   `json:"pin_comment,omitempty"` // 置顶评论
	NextCursor int64                  `json:"next_cursor"`
	HasNext    bool                   `json:"has_next"`
}

// 删除评论
type DelReq struct {
	CommentId int64        `json:"comment_id"`
	Oid       model.NoteId `json:"oid"` // 被评论对象id
}

type PinAction int8

const (
	PinActionUnpin = 0
	PinActionPin   = 1
)

// 置顶评论
type PinReq struct {
	Oid       model.NoteId `json:"oid"`
	CommentId int64        `json:"comment_id"`
	Action    PinAction    `json:"action"`
}

func (r *PinReq) Validate() error {
	if r.Action != PinActionUnpin && r.Action != PinActionPin {
		return xerror.ErrArgs.Msg("不支持的置顶操作")
	}

	return nil
}

type ThumbAction uint8

const (
	ThumbActionUndo ThumbAction = ThumbAction(commentv1.CommentAction_REPLY_ACTION_UNDO) // 取消 0
	ThumbActionDo   ThumbAction = ThumbAction(commentv1.CommentAction_REPLY_ACTION_DO)   // 执行 1
)

type thumbActionChecker struct{}

func (c thumbActionChecker) check(action ThumbAction) error {
	if action != ThumbActionUndo && action != ThumbActionDo {
		return xerror.ErrArgs.Msg("不支持的操作")
	}

	return nil
}

// 点赞评论/取消点赞评论
type ThumbUpReq struct {
	thumbActionChecker
	CommentId int64       `json:"comment_id"`
	Action    ThumbAction `json:"action"`
}

func (r *ThumbUpReq) Validate() error {
	return r.check(r.Action)
}

// 点踩评论/取消点踩评论
type ThumbDownReq struct {
	thumbActionChecker
	CommentId int64       `json:"comment_id"`
	Action    ThumbAction `json:"action"`
}

func (r *ThumbDownReq) Validate() error {
	return r.check(r.Action)
}

type GetLikeCountReq struct {
	CommentId int64 `form:"comment_id"`
}

func (r *GetLikeCountReq) Validate() error {
	if r.CommentId <= 0 {
		return xerror.ErrArgs.Msg("评论不存在")
	}

	return nil
}

type GetLikeCountRes struct {
	Comment int64 `json:"comment_id"`
	Likes   int64 `json:"likes"`
}

type UploadImagesReq struct {
	Count int32 `form:"count"`
}

const (
	maxNumCommentImages = 9
)

func (r *UploadImagesReq) Validate() error {
	if r.Count <= 0 {
		return xerror.ErrArgs.Msg("参数错误")
	}

	if r.Count > maxNumCommentImages {
		return xerror.ErrArgs.Msg(fmt.Sprintf("最多上传%d张图片", maxNumCommentImages))
	}

	return nil
}

type UploadImagesRes = commentv1.UploadCommentImagesResponse
