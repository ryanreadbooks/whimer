package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	RelationErrCode = xerror.BizRelation

	ErrInvalidArgsCode = RelationErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrRelationNilReqCode = ErrInvalidArgsCode + iota
	ErrRelationFollowSelfCode
	ErrRelationUnFollowSelfCode
	ErrRelationAlreadyFollowCode
	ErrRelationUserNotFoundCode
	ErrRelationLockNotHeldCode
	ErrRelationUnSupportedCode
)

const (
	_ = iota

	ErrRelationNotAllowedGetFanListCode = ErrPermissionCode + iota
	ErrRelationNotAllowedGetFollowingListCode
	ErrRelationFollowReachMaxCountCode
	ErrRelationFanListHiddenCode
	ErrRelationFollowingListHiddenCode
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs                       = ErrBizArgs.Msg("参数错误")
	ErrUnSupported                = ErrBizArgs.ErrCode(ErrRelationUnSupportedCode).Msg("不支持的操作")
	ErrInternal                   = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied                 = ErrBizDenied.Msg("操作权限不足")
	ErrNilReq                     = ErrBizArgs.ErrCode(ErrRelationNilReqCode).Msg("请求参数为空")
	ErrFollowSelf                 = ErrBizArgs.ErrCode(ErrRelationFollowSelfCode).Msg("不能关注自己")
	ErrUnFollowSelf               = ErrBizArgs.ErrCode(ErrRelationUnFollowSelfCode).Msg("不能取消关注自己")
	ErrAlreadyFollow              = ErrBizArgs.ErrCode(ErrRelationAlreadyFollowCode).Msg("无需重复关注")
	ErrNotAllowedGetFanList       = ErrBizDenied.ErrCode(ErrRelationNotAllowedGetFanListCode).Msg("不能获取他人的粉丝列表")
	ErrNotAllowedGetFollowingList = ErrBizDenied.ErrCode(ErrRelationNotAllowedGetFollowingListCode).Msg("不能获取他人的关注列表")
	ErrFollowReachMaxCount        = ErrBizDenied.ErrCode(ErrRelationFollowReachMaxCountCode).Msg("关注已达上限")
	ErrUserNotFound               = ErrBizArgs.ErrCode(ErrRelationUserNotFoundCode).Msg("用户不存在")
	ErrLockNotHeld                = ErrBizArgs.ErrCode(ErrRelationLockNotHeldCode).Msg("操作过快")
	ErrFanListHidden              = ErrBizDenied.ErrCode(ErrRelationFanListHiddenCode).Msg("用户隐藏了粉丝列表")
	ErrFollowingListHidden        = ErrBizDenied.ErrCode(ErrRelationFollowingListHiddenCode).Msg("用户隐藏了关注列表")
)
