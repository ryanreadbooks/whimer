package global

import "github.com/ryanreadbooks/whimer/misc/errorx"

const (
	NoteErrCode = errorx.BizNote
)

// 业务错误定义
var (
	ErrBizNoteArgs     = errorx.ErrInvalidArgs.ErrCode(NoteErrCode)
	ErrBizNoteInternal = errorx.ErrInternal.ErrCode(NoteErrCode)
	ErrBizNoteDenied   = errorx.ErrPermission.ErrCode(NoteErrCode)
	ErrNotFound        = errorx.ErrNotFound.ErrCode(NoteErrCode)

	ErrArgs       = ErrBizNoteArgs.Msg("笔记参数错误")
	ErrInternal   = ErrBizNoteInternal.Msg("笔记服务错误, 请稍后重试")
	ErrPermDenied = ErrBizNoteDenied.Msg("你的操作权限不足")

	// 通用错误
	ErrNilReq              = ErrArgs.Msg("请求参数为空")
	ErrUnsupportedResource = ErrArgs.Msg("不支持的资源类型")

	// 笔记操作失败
	ErrInsertNoteFail   = ErrInternal.Msg("添加笔记失败")
	ErrNoteNotFound     = ErrNotFound.Msg("笔记不存在")
	ErrUpdateNoteFail   = ErrInternal.Msg("更新笔记失败")
	ErrDeleteNoteFail   = ErrInternal.Msg("删除笔记失败")
	ErrGetNoteFail      = ErrInternal.Msg("获取笔记失败")
	ErrGetNoteLikesFail = ErrInternal.Msg("获取点赞数据失败")
)
