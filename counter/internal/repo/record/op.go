package record

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
)

// all sqls here
const (
	fields = "biz_code,uid,oid,act,ctime,mtime"

	sqlUpdate = "UPDATE counter_record SET act=?, mtime=? WHERE uid=? AND oid=? AND biz_code=?"
	sqlCount  = "SELECT COUNT(*) FROM counter_record WHERE oid=? AND biz_code=? AND act=?"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO counter_record(%s) VALUES(?,?,?,?,?,?)", fields)
	sqlInUpd  = fmt.Sprintf("INSERT INTO counter_record(%s) VALUES(?,?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE act=val.act, mtime=val.mtime", fields)
	sqlFind = fmt.Sprintf("SELECT %s FROM counter_record WHERE uid=? AND oid=? AND biz_code=?", fields)
)

func (r *Repo) InsertUpdate(ctx context.Context, data *Model) error {
	if data.Ctime <= 0 {
		data.Ctime = time.Now().Unix()
	}

	if data.Mtime <= 0 {
		data.Mtime = data.Ctime
	}

	_, err := r.db.ExecCtx(ctx, sqlInUpd,
		data.BizCode,
		data.Uid,
		data.Oid,
		data.Act,
		data.Ctime,
		data.Mtime)

	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (r *Repo) Insert(ctx context.Context, data *Model) error {
	if data.Ctime <= 0 {
		data.Ctime = time.Now().Unix()
	}

	if data.Mtime <= 0 {
		data.Mtime = data.Ctime
	}

	_, err := r.db.ExecCtx(ctx, sqlInsert,
		data.BizCode,
		data.Uid,
		data.Oid,
		data.Act,
		data.Ctime,
		data.Mtime)

	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (r *Repo) Update(ctx context.Context, data *Model) error {
	_, err := r.db.ExecCtx(ctx,
		sqlUpdate,
		data.Act,
		time.Now().Unix(),
		data.Uid,
		data.Oid,
		data.BizCode)
	return xsql.ConvertError(err)
}

func (r *Repo) Find(ctx context.Context, uid, oid uint64, biz int) (*Model, error) {
	var ret Model
	err := r.db.QueryRowCtx(ctx, &ret, sqlFind, uid, oid, biz)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &ret, nil
}

func (r *Repo) Count(ctx context.Context, oid uint64, biz int) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCount, oid, biz, ActDo)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}
