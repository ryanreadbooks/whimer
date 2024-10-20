package summary

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	fields = "biz_code,oid,cnt,ctime,mtime"

	sqlGet  = "SELECT cnt FROM counter_summary WHERE oid=? AND biz_code=?"
	sqlGets = "SELECT biz_code,oid,cnt FROM counter_summary WHERE oid IN (%s) AND biz_code IN (%s)"

	sqlIncr = "UPDATE counter_summary SET cnt=cnt+1 WHERE oid=? and biz_code=?"
	sqlDecr = "UPDATE counter_summary SET cnt=cnt-1 WHERE oid=? and biz_code=?"

	sqlInsertIncr = "INSERT INTO counter_summary(%s) VALUES (?,?,?,?,?) AS val " +
		"ON DUPLICATE KEY UPDATE counter_summary.cnt=counter_summary.cnt+1, mtime=val.mtime"
	sqlInsertDecr = "INSERT INTO counter_summary(%s) VALUES (?,?,?,?,?) AS val " +
		"ON DUPLICATE KEY UPDATE counter_summary.cnt=counter_summary.cnt-1, mtime=val.mtime"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO counter_summary(%s) VALUES(?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE cnt=val.cnt, mtime=val.mtime", fields)

	sqlBatchInsert = fmt.Sprintf("INSERT INTO counter_summary(%s) VALUES %%s AS val "+
		"ON DUPLICATE KEY UPDATE cnt=val.cnt, mtime=val.mtime", fields)
)

func (r *Repo) Insert(ctx context.Context, data *Model) error {
	if data.Ctime <= 0 {
		data.Ctime = time.Now().Unix()
	}

	if data.Mtime <= 0 {
		data.Mtime = data.Ctime
	}

	_, err := r.db.ExecCtx(ctx, sqlInsert,
		data.BizCode,
		data.Oid,
		data.Cnt,
		data.Ctime,
		data.Mtime)

	return xsql.ConvertError(err)
}

func modelAsInsertSql(data *Model, builder *strings.Builder) {
	builder.WriteByte('(')
	builder.WriteString(strconv.Itoa(int(data.BizCode)))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatUint(data.Oid, 10))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatUint(data.Cnt, 10))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatInt(data.Ctime, 10))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatInt(data.Mtime, 10))
	builder.WriteByte(')')
}

func (r *Repo) BatchInsert(ctx context.Context, datas []*Model) error {
	if len(datas) == 0 {
		return nil
	}

	now := time.Now().Unix()
	var builder strings.Builder
	if datas[0].Ctime <= 0 {
		datas[0].Ctime = now
	}
	if datas[0].Mtime <= 0 {
		datas[0].Mtime = datas[0].Ctime
	}
	modelAsInsertSql(datas[0], &builder)
	for i := 1; i < len(datas); i++ {
		data := datas[i]
		if data.Ctime <= 0 {
			data.Ctime = now
		}
		if data.Mtime <= 0 {
			data.Mtime = data.Ctime
		}
		builder.WriteByte(',')
		modelAsInsertSql(data, &builder)
	}
	sql := fmt.Sprintf(sqlBatchInsert, builder.String())
	_, err := r.db.ExecCtx(ctx, sql)

	return xsql.ConvertError(err)
}

func (r *Repo) Get(ctx context.Context, biz int, oid uint64) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlGet, oid, biz)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) Gets(ctx context.Context, keys PrimaryKeyList) (map[PrimaryKey]uint64, error) {
	oids := slices.JoinInts(slices.Uniq(keys.Oids()))
	bizs := slices.JoinInts(slices.Uniq(keys.BizCodes()))
	sql := fmt.Sprintf(sqlGets, oids, bizs)

	var data = make([]*GetsResult, 0, len(keys))
	err := r.db.QueryRowsCtx(ctx, &data, sql)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	ret := make(map[PrimaryKey]uint64, len(keys))
	for _, item := range data {
		k := PrimaryKey{BizCode: item.BizCode, Oid: item.Oid}
		ret[k] = item.Cnt
	}

	return ret, nil
}

func (r *Repo) Incr(ctx context.Context, biz int, oid uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlIncr, oid, biz)
	if mysqlerr, ok := err.(*mysql.MySQLError); ok {
		if xsql.SQLStateEqual(mysqlerr.SQLState, xsql.SQLStateOutOfRange) {
			return xsql.ErrOutOfRange
		}
	}
	return xsql.ConvertError(err)
}

func (r *Repo) Decr(ctx context.Context, biz int, oid uint64) error {
	_, err := r.db.ExecCtx(ctx, sqlDecr, oid, biz)
	if mysqlerr, ok := err.(*mysql.MySQLError); ok {
		if xsql.SQLStateEqual(mysqlerr.SQLState, xsql.SQLStateOutOfRange) {
			return xsql.ErrOutOfRange
		}
	}
	return xsql.ConvertError(err)
}

// "biz_code,oid,cnt,ctime,mtime"
func (r *Repo) InsertOrIncr(ctx context.Context, biz int, oid uint64) error {
	now := time.Now().Unix()
	_, err := r.db.ExecCtx(ctx, fmt.Sprintf(sqlInsertIncr, fields),
		biz,
		oid,
		1,
		now,
		now,
	)
	if mysqlerr, ok := err.(*mysql.MySQLError); ok {
		if xsql.SQLStateEqual(mysqlerr.SQLState, xsql.SQLStateOutOfRange) {
			return xsql.ErrOutOfRange
		}
	}
	return xsql.ConvertError(err)
}

func (r *Repo) InsertOrDecr(ctx context.Context, biz int, oid uint64) error {
	now := time.Now().Unix()
	_, err := r.db.ExecCtx(ctx, fmt.Sprintf(sqlInsertDecr, fields),
		biz,
		oid,
		1,
		now,
		now,
	)
	if mysqlerr, ok := err.(*mysql.MySQLError); ok {
		if xsql.SQLStateEqual(mysqlerr.SQLState, xsql.SQLStateOutOfRange) {
			return xsql.ErrOutOfRange
		}
	}
	return xsql.ConvertError(err)
}
