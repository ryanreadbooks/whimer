package dao

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type LockDao struct {
	db *xsql.DB
}

func NewLockDao(db *xsql.DB) *LockDao {
	return &LockDao{
		db: db,
	}
}

// TryAcquireLock 尝试获取锁
func (d *LockDao) TryAcquireLock(ctx context.Context, lockKey, lockVal, heldBy string, ttl int64) (bool, error) {
	// 使用 INSERT IGNORE 来尝试插入锁记录
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(lockPOTableName)
	ib.Cols(lockPOFields...)
	ib.Values([]any{
		0, // id 为自增字段，设为0
		lockKey,
		lockVal,
		heldBy,
		time.Now().Unix() + ttl, // expire time
		time.Now().Unix(),       // create time
	}...)

	sql, args := ib.Build()
	result, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return false, xsql.ConvertError(err)
	}

	// 检查是否插入成功
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, xsql.ConvertError(err)
	}

	// 如果影响行数大于0，说明获取锁成功
	return rowsAffected > 0, nil
}

// ReleaseLock 释放锁
func (d *LockDao) ReleaseLock(ctx context.Context, lockKey, heldBy string) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(lockPOTableName)
	db.Where(
		db.Equal("lock_key", lockKey),
		db.Equal("held_by", heldBy),
	)

	sql, args := db.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// RenewLock 续约锁
func (d *LockDao) RenewLock(ctx context.Context, lockKey, heldBy string, ttl int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(lockPOTableName)
	ub.Set(
		ub.Assign("expire", time.Now().Unix()+ttl),
	)
	ub.Where(
		ub.Equal("lock_key", lockKey),
		ub.Equal("held_by", heldBy),
		ub.GreaterThan("expire", time.Now().Unix()), // 确保锁还没有过期
	)

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// GetLockInfo 获取锁信息
func (d *LockDao) GetLockInfo(ctx context.Context, lockKey string) (*LockPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(lockPOFields...)
	sb.From(lockPOTableName)
	sb.Where(
		sb.Equal("lock_key", lockKey),
		sb.GreaterThan("expire", time.Now().Unix()), // 只返回未过期的锁
	)

	sql, args := sb.Build()
	var po LockPO
	err := d.db.QueryRowCtx(ctx, &po, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &po, nil
}

// CleanupExpiredLocks 清理过期锁
func (d *LockDao) CleanupExpiredLocks(ctx context.Context) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(lockPOTableName)
	db.Where(
		db.LessEqualThan("expire", time.Now().Unix()),
	)

	sql, args := db.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

