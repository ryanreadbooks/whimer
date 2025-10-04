package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	CommentErrCode = xerror.BizComment
)

// 业务错误定义
var (
	ErrBizCommentArgs     = xerror.ErrInvalidArgs.ErrCode(CommentErrCode)
	ErrBizCommentInternal = xerror.ErrInternal.ErrCode(CommentErrCode)
	ErrBizCommentDenied   = xerror.ErrPermission.ErrCode(CommentErrCode)
	ErrNotFound           = xerror.ErrNotFound.ErrCode(CommentErrCode)

	ErrArgs       = ErrBizCommentArgs.Msg("参数错误")
	ErrInternal   = ErrBizCommentInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied = ErrBizCommentDenied.Msg("操作权限不足")

	ErrNilReq               = ErrArgs.Msg("请求参数为空")
	ErrNoNote               = ErrNotFound.Msg("内容不存在")
	ErrUnsupportedType      = ErrArgs.Msg("内容类型不支持")
	ErrObjectIdEmpty        = ErrArgs.Msg("对象id为空")
	ErrReplyUidEmpty        = ErrArgs.Msg("回复用户id为空")
	ErrContentTooShort      = ErrArgs.Msg("评论内容太短")
	ErrContentTooLong       = ErrArgs.Msg("评论内容太长")
	ErrInvalidCommentId     = ErrArgs.Msg("评论id错误")
	ErrCommentNotFound      = ErrNotFound.Msg("评论不存在")
	ErrCommentWrongRelation = ErrArgs.Msg("评论关系错误")
	ErrYouDontOwnThis       = ErrPermDenied.Msg("你不是该评论的作者")
	ErrRootCommentIsNotRoot = ErrArgs.Msg("指定根评论并非根评论")
	ErrInvalidImageCount    = ErrArgs.Msg("评论图片数量错误")

	ErrPinFailInternal        = ErrInternal.Msg("置顶操作失败，请稍后重试")
	ErrUnPinFailInternal      = ErrInternal.Msg("取消置顶操作失败，请稍后重试")
	ErrGetPinnedInternal      = ErrInternal.Msg("获取置顶评论失败")
	ErrPinFailNotRoot         = ErrArgs.Msg("不能操作非主评论")
	ErrOidNotMatch            = ErrArgs.Msg("评论对象id不匹配")
	ErrYouCantPinComment      = ErrPermDenied.Msg("你无权置顶评论")
	ErrCountCommentInternal   = ErrArgs.Msg("获取评论数量失败")
	ErrGetCommentLikeCount    = ErrArgs.Msg("获取评论点赞失败")
	ErrGetCommentDislikeCount = ErrArgs.Msg("获取评论点踩失败")
	ErrUnsupportedAction      = ErrArgs.Msg("不支持的操作")
	ErrNoPinComment           = ErrArgs.Msg("无置顶评论")
)
