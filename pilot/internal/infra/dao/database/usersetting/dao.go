package usersetting

import (
	"context"
	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type Dao struct {
	db *xsql.DB
}

func NewDao(db *xsql.DB) *Dao {
	return &Dao{
		db: db,
	}
}

func (d *Dao) Upsert(ctx context.Context, po *UserSettingPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(userSettingPOTableName)
	ib.Cols(userSettingPOFields...)
	ib.Values(po.Values()...)
	extra := "ON DUPLICATE KEY UPDATE utime=VALUES(utime), flags=VALUES(flags)"
	ib.SQL(extra)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)

	return xsql.ConvertError(err)
}

func (d *Dao) GetByUid(ctx context.Context, uid int64, forUpdate bool) (*UserSettingPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(userSettingPOFields...)
	sb.From(userSettingPOTableName)
	sb.Where(sb.EQ("uid", uid))
	if forUpdate {
		sb.SQL("FOR UPDATE")
	}

	sql, args := sb.Build()

	var po UserSettingPO
	err := d.db.QueryRowCtx(ctx, &po, sql, args...)
	return &po, xsql.ConvertError(err)
}
