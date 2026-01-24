package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrUnsupportedVisibility = xerror.ErrArgs.Msg("不支持的可见范围")
	ErrWrongTitleLength      = xerror.ErrArgs.Msg("标题长度错误")
	ErrWrongDescLength       = xerror.ErrArgs.Msg("简介超长")
	ErrWrongNoteType         = xerror.ErrArgs.Msg("不支持资源类型")
	ErrNilVideoParam         = xerror.ErrArgs.Msg("视频参数为空")
	ErrEmptyVideoFileId      = xerror.ErrArgs.Msg("未指定视频资源")
	ErrEmptyCoverFileId      = xerror.ErrArgs.Msg("未指定封面")
	ErrNilImageParam         = xerror.ErrArgs.Msg("图片参数为空")
	ErrEmptyImageFileId      = xerror.ErrArgs.Msg("上传图片无标识")
	ErrEmptyImageInfo        = xerror.ErrArgs.Msg("上传图片未指定图片信息")
	ErrTagCountExceed        = xerror.ErrArgs.Msg("标签超出限制")
	ErrNilArg                = xerror.ErrArgs.Msg("参数为空")
	ErrResourceNotFound      = xerror.ErrArgs.Msg("资源不存在")
	ErrNoteNotFound          = xerror.ErrArgs.Msg("笔记不存在")
)
