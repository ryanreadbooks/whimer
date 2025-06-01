package record

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

// all sqls here
const (
	fields    = "biz_code,uid,oid,act,ctime,mtime"
	allFields = "id,biz_code,uid,oid,act,ctime,mtime"

	sqlUpdate     = "UPDATE counter_record SET act=?, mtime=? WHERE uid=? AND oid=? AND biz_code=?"
	sqlCount      = "SELECT COUNT(*) FROM counter_record WHERE oid=? AND biz_code=? AND act=?"
	sqlCountAll   = "SELECT COUNT(*) FROM counter_record WHERE act=?"
	sqlPageGet    = "SELECT %s FROM counter_record WHERE id>=? AND act=? LIMIT ?"
	sqlGetSummary = "SELECT biz_code,oid,count(1) cnt FROM counter_record WHERE act=? GROUP BY biz_code,oid"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO counter_record(%s) VALUES(?,?,?,?,?,?)", fields)
	sqlInUpd  = fmt.Sprintf("INSERT INTO counter_record(%s) VALUES(?,?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE act=val.act, mtime=val.mtime", fields)
	sqlFind      = fmt.Sprintf("SELECT %s FROM counter_record WHERE uid=? AND oid=? AND biz_code=?", allFields)
	sqlBatchFind = fmt.Sprintf("SELECT DISTINCT %s FROM counter_record WHERE uid IN (%%s) AND oid IN (%%s) AND biz_code=?", allFields)
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

func (r *Repo) Find(ctx context.Context, uid int64, oid uint64, biz int) (*Model, error) {
	var ret Model
	err := r.db.QueryRowCtx(ctx, &ret, sqlFind, uid, oid, biz)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &ret, nil
}

func (r *Repo) BatchFind(ctx context.Context, uidOids map[int64][]uint64, biz int) ([]Model, error) {
	var batchRes []Model
	// 分批操作
	err := maps.BatchExec(uidOids, 200, func(target map[int64][]uint64) error {
		uids, oids := maps.All(target)
		var allOids []uint64 = oids[0]
		for i := 1; i < len(oids); i++ {
			allOids = slices.Concat(allOids, oids[i])
		}

		var ret = make([]Model, 0, len(uids)*len(allOids)) // we should strictly limit the length of them
		query := fmt.Sprintf(sqlBatchFind, slices.JoinInts(uids), slices.JoinInts(allOids))
		err := r.db.QueryRowsCtx(ctx, &ret, query, biz)
		if err != nil {
			return err
		}

		batchRes = append(batchRes, ret...)
		return nil
	})

	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return batchRes, nil
}

func (r *Repo) Count(ctx context.Context, oid uint64, biz int) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCount, oid, biz, ActDo)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) PageGet(ctx context.Context, id uint64, act int, limit uint64) ([]*Model, error) {
	var res = make([]*Model, 0)
	err := r.db.QueryRowsCtx(ctx, &res, fmt.Sprintf(sqlPageGet, allFields), id, act, limit)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

func (r *Repo) CountAll(ctx context.Context) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountAll, ActDo)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) GetSummary(ctx context.Context, act int) ([]*Summary, error) {
	var summaries []*Summary
	err := r.db.QueryRowsCtx(ctx, &summaries, sqlGetSummary, act)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return summaries, nil
}
