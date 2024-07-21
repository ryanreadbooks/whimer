package global

import "github.com/ryanreadbooks/whimer/misc/errorx"

const (
	CommentErrCode = errorx.BizComment
)

// 业务错误定义
var (
	ErrBizCommentArgs     = errorx.ErrInvalidArgs.ErrCode(CommentErrCode)
	ErrBizCommentInternal = errorx.ErrInternal.ErrCode(CommentErrCode)
	ErrBizCommentDenied   = errorx.ErrPermission.ErrCode(CommentErrCode)
	ErrNotFound           = errorx.ErrNotFound.ErrCode(CommentErrCode)

	ErrArgs       = ErrBizCommentArgs.Msg("参数错误")
	ErrInternal   = ErrBizCommentInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied = ErrBizCommentDenied.Msg("你的操作权限不足")

	ErrNilReq          = ErrArgs.Msg("请求参数为空")
	ErrUnsupportedType = ErrArgs.Msg("内容类型不支持")
	ErrObjectIdEmpty   = ErrArgs.Msg("对象id为空")
	ErrContentTooShort = ErrArgs.Msg("评论内容太短")
	ErrContentTooLong  = ErrArgs.Msg("评论内容太长")
)
