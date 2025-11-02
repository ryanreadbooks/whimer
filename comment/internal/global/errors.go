package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	CommentErrCode = xerror.BizComment

	ErrInvalidArgsCode = CommentErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrCommentNilReqCode = ErrInvalidArgsCode + iota
	ErrCommentUnsupportedTypeCode
	ErrCommentObjectIdEmptyCode
	ErrCommentReplyUidEmptyCode
	ErrCommentContentTooShortCode
	ErrCommentContentTooLongCode
	ErrCommentInvalidCommentIdCode
	ErrCommentCommentWrongRelationCode
	ErrCommentRootCommentIsNotRootCode
	ErrCommentInvalidImageCountCode
	ErrCommentPinFailNotRootCode
	ErrCommentOidNotMatchCode
	ErrCommentCountCommentInternalCode
	ErrCommentGetCommentLikeCountCode
	ErrCommentGetCommentDislikeCountCode
	ErrCommentUnsupportedActionCode
	ErrCommentNoPinCommentCode
)

const (
	_ = iota

	ErrCommentPinFailInternalCode = ErrInternalCode + iota
	ErrCommentUnPinFailInternalCode
	ErrCommentGetPinnedInternalCode
)

const (
	_ = iota

	ErrCommentYouDontOwnThisCode = ErrPermissionCode + iota
	ErrCommentYouCantPinCommentCode
)

const (
	_ = iota

	ErrCommentNoNoteCode = ErrNotFoundCode + iota
	ErrCommentCommentNotFoundCode
)

// 业务错误定义
var (
	ErrBizCommentArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizCommentInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizCommentDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound           = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs       = ErrBizCommentArgs.Msg("参数错误")
	ErrInternal   = ErrBizCommentInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied = ErrBizCommentDenied.Msg("操作权限不足")

	ErrNilReq               = ErrBizCommentArgs.ErrCode(ErrCommentNilReqCode).Msg("请求参数为空")
	ErrNoNote               = ErrNotFound.ErrCode(ErrCommentNoNoteCode).Msg("内容不存在")
	ErrUnsupportedType      = ErrBizCommentArgs.ErrCode(ErrCommentUnsupportedTypeCode).Msg("内容类型不支持")
	ErrObjectIdEmpty        = ErrBizCommentArgs.ErrCode(ErrCommentObjectIdEmptyCode).Msg("对象id为空")
	ErrReplyUidEmpty        = ErrBizCommentArgs.ErrCode(ErrCommentReplyUidEmptyCode).Msg("回复用户id为空")
	ErrContentTooShort      = ErrBizCommentArgs.ErrCode(ErrCommentContentTooShortCode).Msg("评论内容太短")
	ErrContentTooLong       = ErrBizCommentArgs.ErrCode(ErrCommentContentTooLongCode).Msg("评论内容太长")
	ErrInvalidCommentId     = ErrBizCommentArgs.ErrCode(ErrCommentInvalidCommentIdCode).Msg("评论id错误")
	ErrCommentNotFound      = ErrNotFound.ErrCode(ErrCommentCommentNotFoundCode).Msg("评论不存在")
	ErrCommentWrongRelation = ErrBizCommentArgs.ErrCode(ErrCommentCommentWrongRelationCode).Msg("评论关系错误")
	ErrYouDontOwnThis       = ErrBizCommentDenied.ErrCode(ErrCommentYouDontOwnThisCode).Msg("你不是该评论的作者")
	ErrRootCommentIsNotRoot = ErrBizCommentArgs.ErrCode(ErrCommentRootCommentIsNotRootCode).Msg("指定根评论并非根评论")
	ErrInvalidImageCount    = ErrBizCommentArgs.ErrCode(ErrCommentInvalidImageCountCode).Msg("评论图片数量错误")

	ErrPinFailInternal        = ErrBizCommentInternal.ErrCode(ErrCommentPinFailInternalCode).Msg("置顶操作失败，请稍后重试")
	ErrUnPinFailInternal      = ErrBizCommentInternal.ErrCode(ErrCommentUnPinFailInternalCode).Msg("取消置顶操作失败，请稍后重试")
	ErrGetPinnedInternal      = ErrBizCommentInternal.ErrCode(ErrCommentGetPinnedInternalCode).Msg("获取置顶评论失败")
	ErrPinFailNotRoot         = ErrBizCommentArgs.ErrCode(ErrCommentPinFailNotRootCode).Msg("不能操作非主评论")
	ErrOidNotMatch            = ErrBizCommentArgs.ErrCode(ErrCommentOidNotMatchCode).Msg("评论对象id不匹配")
	ErrYouCantPinComment      = ErrBizCommentDenied.ErrCode(ErrCommentYouCantPinCommentCode).Msg("你无权置顶评论")
	ErrCountCommentInternal   = ErrBizCommentArgs.ErrCode(ErrCommentCountCommentInternalCode).Msg("获取评论数量失败")
	ErrGetCommentLikeCount    = ErrBizCommentArgs.ErrCode(ErrCommentGetCommentLikeCountCode).Msg("获取评论点赞失败")
	ErrGetCommentDislikeCount = ErrBizCommentArgs.ErrCode(ErrCommentGetCommentDislikeCountCode).Msg("获取评论点踩失败")
	ErrUnsupportedAction      = ErrBizCommentArgs.ErrCode(ErrCommentUnsupportedActionCode).Msg("不支持的操作")
	ErrNoPinComment           = ErrBizCommentArgs.ErrCode(ErrCommentNoPinCommentCode).Msg("无置顶评论")
)
