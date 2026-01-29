package dto

import "github.com/ryanreadbooks/whimer/misc/xerror"

func CheckImageFormat(format string) error {
	switch format {
	case "image/jpg", "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return xerror.ErrArgs.Msg("不支持的图片格式")
	}
}
