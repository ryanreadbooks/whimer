package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	RelationErrCode = 60000
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(RelationErrCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(RelationErrCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(RelationErrCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(RelationErrCode)

	ErrArgs                       = ErrBizArgs.Msg("参数错误")
	ErrUnSupported                = ErrBizArgs.Msg("不支持的操作")
	ErrInternal                   = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied                 = ErrBizDenied.Msg("操作权限不足")
	ErrNilReq                     = ErrArgs.Msg("请求参数为空")
	ErrFollowSelf                 = ErrArgs.Msg("不能关注自己")
	ErrUnFollowSelf               = ErrArgs.Msg("不能取消关注自己")
	ErrAlreadyFollow              = ErrArgs.Msg("无需重复关注")
	ErrNotAllowedGetFanList       = ErrPermDenied.Msg("不能获取他人的粉丝列表")
	ErrNotAllowedGetFollowingList = ErrPermDenied.Msg("不能获取他人的关注列表")
)
