package model

import "github.com/ryanreadbooks/whimer/msger/internal/global"

func CheckImageFormat(format string) error {
	switch format {
	case "image/jpg", "image/jpeg", "image/png", "image/webp", "image/gif":
		return nil
	default:
		return global.ErrArgs.Msg("unsupported image format")
	}
}
