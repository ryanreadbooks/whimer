package global

import "github.com/ryanreadbooks/whimer/misc/errorx"

const (
	NoteErrCode = 100100000
)

// 业务错误定义
var (
	ErrBizNoteArgs     = errorx.ErrInvalidArgs.ErrCode(NoteErrCode)
	ErrBizNoteInternal = errorx.ErrInternal.ErrCode(NoteErrCode)
	ErrBizNoteDenied   = errorx.ErrPermission.ErrCode(NoteErrCode)

	ErrNotFound   = errorx.ErrNotFound.ErrCode(NoteErrCode).Msg("资源不存在")
	ErrArgs       = ErrBizNoteArgs.Msg("笔记参数错误")
	ErrInternal   = ErrBizNoteInternal.Msg("笔记服务错误, 请稍后重试")
	ErrPermDenied = ErrBizNoteDenied.Msg("你的操作权限不足")

	// 笔记操作失败
	ErrInsertNoteFail     = ErrInternal.Msg("添加笔记失败")
	ErrUpdateNoteNotFound = ErrNotFound.Msg("笔记不存在")
	ErrUpdateNoteFail     = ErrInternal.Msg("更新笔记失败")
)
