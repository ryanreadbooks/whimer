package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	NoteErrCode = xerror.BizNote

	// Note error groups

	ErrInvalidArgsCode = NoteErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrNoteNilReqCode = ErrInvalidArgsCode + iota
	ErrNoteUnsupportedResourceCode
)

const (
	_ = iota

	ErrNoteInsertNoteFailCode = ErrInternalCode + iota
	ErrNoteUpdateNoteFailCode
	ErrNoteDeleteNoteFailCode
	ErrNoteGetNoteFailCode
	ErrNoteGetNoteLikesFailCode
)

const (
	_ = iota

	ErrNoteNoteNotFoundCode = ErrNotFoundCode + iota
)

const (
	_ = iota

	ErrNoteNotNoteOwnerCode = ErrPermissionCode + iota
	ErrNoteNoteNotPublicCode
)

// 业务错误定义
var (
	ErrBizNoteArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizNoteInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizNoteDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound        = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs       = ErrBizNoteArgs.Msg("笔记参数错误")
	ErrInternal   = ErrBizNoteInternal.Msg("笔记服务错误, 请稍后重试")
	ErrPermDenied = ErrBizNoteDenied.Msg("你的操作权限不足")

	// 通用错误
	ErrNilReq              = ErrBizNoteArgs.ErrCode(ErrNoteNilReqCode).Msg("请求参数为空")
	ErrUnsupportedResource = ErrBizNoteArgs.ErrCode(ErrNoteUnsupportedResourceCode).Msg("不支持的资源类型")

	// 笔记操作失败
	ErrInsertNoteFail   = ErrBizNoteInternal.ErrCode(ErrNoteInsertNoteFailCode).Msg("添加笔记失败")
	ErrNoteNotFound     = ErrNotFound.ErrCode(ErrNoteNoteNotFoundCode).Msg("笔记不存在")
	ErrUpdateNoteFail   = ErrBizNoteInternal.ErrCode(ErrNoteUpdateNoteFailCode).Msg("更新笔记失败")
	ErrDeleteNoteFail   = ErrBizNoteInternal.ErrCode(ErrNoteDeleteNoteFailCode).Msg("删除笔记失败")
	ErrGetNoteFail      = ErrBizNoteInternal.ErrCode(ErrNoteGetNoteFailCode).Msg("获取笔记失败")
	ErrGetNoteLikesFail = ErrBizNoteInternal.ErrCode(ErrNoteGetNoteLikesFailCode).Msg("获取点赞数据失败")
	ErrNotNoteOwner     = ErrBizNoteDenied.ErrCode(ErrNoteNotNoteOwnerCode).Msg("你不拥有该笔记")
	ErrNoteNotPublic    = ErrBizNoteDenied.ErrCode(ErrNoteNoteNotPublicCode).Msg("笔记非公开")
)
