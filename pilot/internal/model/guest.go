package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
)

func IsGuestFromCtx(ctx context.Context) bool {
	return IsGuest(metadata.Uid(ctx))
}

func IsGuest(uid int64) bool {
	return uid == 0
}