package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrCommentNotFound    = xerror.ErrNotFound.Msg("评论不存在")
	ErrNoteNotFound       = xerror.ErrNotFound.Msg("笔记不存在")
	ErrInvalidAction      = xerror.ErrInvalidArgs.Msg("不支持的操作")
	ErrContentTooLong     = xerror.ErrInvalidArgs.Msg("评论内容太长")
	ErrContentEmpty       = xerror.ErrInvalidArgs.Msg("评论内容为空")
	ErrNoCommentImage     = xerror.ErrInvalidArgs.Msg("无评论图片")
	ErrTooManyImages      = xerror.ErrInvalidArgs.Msg("最多支持9张评论图片")
	ErrInvalidStoreKey    = xerror.ErrInvalidArgs.Msg("非法storeKey")
	ErrMissingImageInfo   = xerror.ErrInvalidArgs.Msg("上传图片未指定图片信息")
	ErrInvalidCommentType = xerror.ErrInvalidArgs.Msg("不支持的评论类型")
	ErrInvalidParams      = xerror.ErrInvalidArgs.Msg("参数错误")
	ErrTooManyUploadCount = xerror.ErrInvalidArgs.Msg("最多上传9张图片")
)
