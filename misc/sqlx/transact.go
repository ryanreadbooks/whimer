package sqlx

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TransactFunc func(ctx context.Context, s sqlx.Session) error

type AfterInsert func(id, cnt int64)
