package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	MsgerErrCode = xerror.BizMsger

	// Msger error groups (use iota for spacing)

	ErrInvalidArgsCode = MsgerErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrMsgerNilReqCode = ErrInvalidArgsCode + iota
	ErrMsgerLockNotHeldCode
	ErrMsgerChatMsgNilCode
	ErrMsgerChatNotExistCode
	ErrMsgerChatUserEmptyCode
	ErrMsgerChatSenderEmptyCode
	ErrMsgerChatReceiverEmptyCode
	ErrMsgerUserNotInChatCode
	ErrMsgerMsgAlreadyRecalledCode
	ErrMsgerRecallTimeReachedCode
	ErrMsgerMsgNotExistCode
	ErrMsgerUserNotFoundCode
	ErrMsgerUnsupportedMsgTypeCode
	ErrMsgerUnsupportedChatTypeCode
	ErrMsgerEmptyMsgCode
	ErrMsgerSysChatNotExistCode
	ErrMsgerChatNotNormalCode
	ErrMsgerChatInboxNotExistCode
)

const (
	_                     = iota
	ErrMsgerGenChatIdCode = ErrInternalCode + iota
)

const (
	_                         = iota
	ErrMsgerCantRecallMsgCode = ErrPermissionCode + iota
	ErrMsgerSysChatNotYoursCode
)

// 业务错误定义
var (
	ErrBizMsgerArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizMsgerInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizMsgerDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound         = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs       = ErrBizMsgerArgs.Msg("参数错误")
	ErrInternal   = ErrBizMsgerInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied = ErrBizMsgerDenied.Msg("操作权限不足")

	ErrNilReq      = ErrBizMsgerArgs.ErrCode(ErrMsgerNilReqCode).Msg("请求参数为空")
	ErrLockNotHeld = ErrBizMsgerArgs.ErrCode(ErrMsgerLockNotHeldCode).Msg("请稍后重试")

	ErrChatMsgNil          = ErrBizMsgerArgs.ErrCode(ErrMsgerChatMsgNilCode).Msg("发送消息为空")
	ErrChatNotExist        = ErrBizMsgerArgs.ErrCode(ErrMsgerChatNotExistCode).Msg("会话不存在")
	ErrChatUserEmpty       = ErrBizMsgerArgs.ErrCode(ErrMsgerChatUserEmptyCode).Msg("会话用户不存在")
	ErrChatSenderEmpty     = ErrBizMsgerArgs.ErrCode(ErrMsgerChatSenderEmptyCode).Msg("发送者不存在")
	ErrChatReceiverEmpty   = ErrBizMsgerArgs.ErrCode(ErrMsgerChatReceiverEmptyCode).Msg("接收者不存在")
	ErrUserNotInChat       = ErrBizMsgerArgs.ErrCode(ErrMsgerUserNotInChatCode).Msg("用户不在会话中")
	ErrMsgAlreadyRecalled  = ErrBizMsgerArgs.ErrCode(ErrMsgerMsgAlreadyRecalledCode).Msg("消息已被撤回")
	ErrRecallTimeReached   = ErrBizMsgerArgs.ErrCode(ErrMsgerRecallTimeReachedCode).Msg("超过撤回时间")
	ErrMsgNotExist         = ErrBizMsgerArgs.ErrCode(ErrMsgerMsgNotExistCode).Msg("消息不存在")
	ErrCantRecallMsg       = ErrBizMsgerDenied.ErrCode(ErrMsgerCantRecallMsgCode).Msg("无权撤回该消息")
	ErrUserNotFound        = ErrBizMsgerArgs.ErrCode(ErrMsgerUserNotFoundCode).Msg("用户不存在")
	ErrUnsupportedMsgType  = ErrBizMsgerArgs.ErrCode(ErrMsgerUnsupportedMsgTypeCode).Msg("不支持的消息类型")
	ErrUnsupportedChatType = ErrBizMsgerArgs.ErrCode(ErrMsgerUnsupportedChatTypeCode).Msg("不支持的会话类型")
	ErrEmptyMsg            = ErrBizMsgerArgs.ErrCode(ErrMsgerEmptyMsgCode).Msg("消息内容为空")
	ErrSysChatNotExist     = ErrBizMsgerArgs.ErrCode(ErrMsgerSysChatNotExistCode).Msg("系统会话不存在")
	ErrSysChatNotYours     = ErrBizMsgerDenied.ErrCode(ErrMsgerSysChatNotYoursCode).Msg("系统消息归属错误")
	ErrGenChatId           = ErrBizMsgerInternal.ErrCode(ErrMsgerGenChatIdCode).Msg("无法生成会话id")
	ErrChatNotNormal       = ErrBizMsgerArgs.ErrCode(ErrMsgerChatNotNormalCode).Msg("会话状态异常")
	ErrChatInboxNotExist   = ErrBizMsgerArgs.ErrCode(ErrMsgerChatInboxNotExistCode).Msg("信箱不存在")
)
