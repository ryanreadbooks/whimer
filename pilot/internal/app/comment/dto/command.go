package dto

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/errors"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	cmtvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/vo"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
)

const (
	maxContentLen       = 2000
	maxNumCommentImages = 9
)

// CommentImageCommand 评论图片
type CommentImageCommand struct {
	StoreKey string `json:"store_key"`
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	Format   string `json:"format"`
}

// PublishCommentCommand 发布评论命令
type PublishCommentCommand struct {
	CommentType cmtvo.CommentType    `json:"type"`
	Oid         notevo.NoteId        `json:"oid"`
	Content     string               `json:"content"`
	RootId      int64                `json:"root_id,omitempty,optional"`
	ParentId    int64                `json:"parent_id,omitempty,optional"`
	ReplyUid    int64                `json:"reply_uid"`
	Images      []CommentImageCommand       `json:"images,omitempty,optional"`
	AtUsers     commondto.AtUserList `json:"at_users,omitempty,optional"`
}

func (c *PublishCommentCommand) PubOnOidDirectly() bool {
	return c.RootId == 0 && c.ParentId == 0
}

func (c *PublishCommentCommand) Validate() error {
	if c.Oid == 0 {
		return errors.ErrNoteNotFound
	}

	contentLen := utf8.RuneCountInString(c.Content)
	if contentLen > maxContentLen {
		return errors.ErrContentTooLong
	}

	if len(c.AtUsers) > 0 {
		c.AtUsers = c.AtUsers.Filter()
	}

	switch c.CommentType {
	case cmtvo.CommentTypeText:
		if contentLen <= 0 {
			return errors.ErrContentEmpty
		}
	case cmtvo.CommentTypeImageText:
		if len(c.Images) == 0 {
			return errors.ErrNoCommentImage
		}
		if len(c.Images) > maxNumCommentImages {
			return errors.ErrTooManyImages
		}

		for _, img := range c.Images {
			if img.StoreKey == "" {
				return errors.ErrInvalidStoreKey
			}

			if img.Width == 0 || img.Height == 0 || img.Format == "" {
				return errors.ErrMissingImageInfo
			}

			if err := model.CheckImageFormat(img.Format); err != nil {
				return err
			}
		}
	default:
		return errors.ErrInvalidCommentType
	}

	return nil
}

func (c *PublishCommentCommand) ToRepoParams() *repository.AddCommentParams {
	images := make([]cmtvo.CommentImage, 0, len(c.Images))
	for _, img := range c.Images {
		images = append(images, cmtvo.CommentImage{
			StoreKey: img.StoreKey,
			Width:    img.Width,
			Height:   img.Height,
			Format:   img.Format,
		})
	}

	atUsers := make([]cmtvo.AtUser, 0, len(c.AtUsers))
	for _, au := range c.AtUsers {
		atUsers = append(atUsers, cmtvo.AtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	return &repository.AddCommentParams{
		Type:     int32(c.CommentType),
		Oid:      int64(c.Oid),
		Content:  c.Content,
		RootId:   c.RootId,
		ParentId: c.ParentId,
		ReplyUid: c.ReplyUid,
		Images:   images,
		AtUsers:  atUsers,
	}
}

// PublishCommentResult 发布评论结果
type PublishCommentResult struct {
	CommentId int64 `json:"comment_id"`
}

// DeleteCommentCommand 删除评论命令
type DeleteCommentCommand struct {
	CommentId int64         `json:"comment_id"`
	Oid       notevo.NoteId `json:"oid"`
}

func (c *DeleteCommentCommand) Validate() error {
	if c.CommentId <= 0 {
		return errors.ErrCommentNotFound
	}
	if c.Oid <= 0 {
		return errors.ErrNoteNotFound
	}
	return nil
}

// PinCommentCommand 置顶评论命令
type PinCommentCommand struct {
	Oid       notevo.NoteId   `json:"oid"`
	CommentId int64           `json:"comment_id"`
	Action    cmtvo.PinAction `json:"action"`
}

func (c *PinCommentCommand) Validate() error {
	if c.Action != cmtvo.PinActionUnpin && c.Action != cmtvo.PinActionPin {
		return errors.ErrInvalidAction
	}
	return nil
}

// LikeCommentCommand 点赞评论命令
type LikeCommentCommand struct {
	CommentId int64             `json:"comment_id"`
	Action    cmtvo.ThumbAction `json:"action"`
}

func (c *LikeCommentCommand) Validate() error {
	if c.Action != cmtvo.ThumbActionUndo && c.Action != cmtvo.ThumbActionDo {
		return errors.ErrInvalidAction
	}
	return nil
}

// DislikeCommentCommand 点踩评论命令
type DislikeCommentCommand struct {
	CommentId int64             `json:"comment_id"`
	Action    cmtvo.ThumbAction `json:"action"`
}

func (c *DislikeCommentCommand) Validate() error {
	if c.Action != cmtvo.ThumbActionUndo && c.Action != cmtvo.ThumbActionDo {
		return errors.ErrInvalidAction
	}
	return nil
}

// UploadImagesCommand 上传评论图片命令
type UploadImagesCommand struct {
	Count int32 `form:"count"`
}

func (c *UploadImagesCommand) Validate() error {
	if c.Count <= 0 {
		return errors.ErrInvalidParams
	}

	if c.Count > maxNumCommentImages {
		return errors.ErrTooManyUploadCount
	}

	return nil
}

// UploadTicket 上传凭证
type UploadTicket struct {
	StoreKeys   []string `json:"store_keys"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}
