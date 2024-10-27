package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	NoteErrCode = xerror.BizNote
)

// 业务错误定义
var (
	ErrBizNoteArgs     = xerror.ErrInvalidArgs.ErrCode(NoteErrCode)
	ErrBizNoteInternal = xerror.ErrInternal.ErrCode(NoteErrCode)
	ErrBizNoteDenied   = xerror.ErrPermission.ErrCode(NoteErrCode)
	ErrNotFound        = xerror.ErrNotFound.ErrCode(NoteErrCode)

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
	ErrNotNoteOwner     = ErrPermDenied.Msg("你不拥有该笔记")
)
