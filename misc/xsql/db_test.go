package xsql

import (
	"context"
	"testing"
)

func TestTransact(t *testing.T) {
	db := NewFromEnv()

	txFn := func(ctx context.Context) (int64, error) {
		var num int64
		err := db.QueryRowCtx(ctx, &num, "SELECT 1")
		return num, err
	}

	var result int64

	err := db.Transact(t.Context(), func(ctx context.Context) error {
		return db.Transact(ctx, func(ctx context.Context) error {
			num, err := txFn(ctx)
			result = num
			return err
		})
	})

	t.Log(result)
	t.Log(err)
}
