package record

import (
	"context"
	"fmt"
	"time"

	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	slices "github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type SortOrder string

const (
	Asc  SortOrder = "ASC"
	Desc SortOrder = "DESC"
)

// sqls here
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

func (r *Repo) InsertUpdate(ctx context.Context, data *Record) error {
	now := time.Now().Unix()
	if data.Ctime <= 0 {
		data.Ctime = time.Now().Unix()
	}

	data.Mtime = now

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

func (r *Repo) Insert(ctx context.Context, data *Record) error {
	now := time.Now().Unix()
	if data.Ctime <= 0 {
		data.Ctime = now
	}
	data.Mtime = now

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

func (r *Repo) Update(ctx context.Context, data *Record) error {
	_, err := r.db.ExecCtx(ctx,
		sqlUpdate,
		data.Act,
		time.Now().Unix(),
		data.Uid,
		data.Oid,
		data.BizCode)
	return xsql.ConvertError(err)
}

func (r *Repo) Find(ctx context.Context, uid int64, oid int64, biz int32) (*Record, error) {
	var ret Record
	err := r.db.QueryRowCtx(ctx, &ret, sqlFind, uid, oid, biz)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &ret, nil
}

func (r *Repo) BatchFind(ctx context.Context, uidOids map[int64][]int64, biz int32) ([]Record, error) {
	var batchRes []Record
	// 分批操作
	err := maps.BatchExec(uidOids, 200, func(target map[int64][]int64) error {
		uids, oids := maps.All(target)
		var allOids []int64 = oids[0]
		for i := 1; i < len(oids); i++ {
			allOids = slices.Concat(allOids, oids[i])
		}

		var ret = make([]Record, 0, len(uids)*len(allOids)) // we should strictly limit the length of them
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

func (r *Repo) Count(ctx context.Context, oid int64, biz int32) (int64, error) {
	var cnt int64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCount, oid, biz, ActDo)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return cnt, nil
}

func (r *Repo) PageGet(ctx context.Context, id int64, act int, limit int64) ([]*Record, error) {
	var res = make([]*Record, 0)
	err := r.db.QueryRowsCtx(ctx, &res, fmt.Sprintf(sqlPageGet, allFields), id, act, limit)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return res, nil
}

func (r *Repo) CountAll(ctx context.Context) (int64, error) {
	var cnt int64
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

type PageGetByUidOrderByMtimeParam struct {
	Uid   int64
	Count int32
	Order SortOrder
}

func handlePageGetByUidOrderByMtimeParam(p *PageGetByUidOrderByMtimeParam) {
	if p.Count == 0 {
		p.Count = 20
	}
	if p.Order == "" {
		p.Order = Desc
	}
}

var sqlPageGetByUidOrderByMtime = fmt.Sprintf(
	"SELECT %s FROM counter_record WHERE uid=? AND biz_code=? AND act=? ORDER BY %%s LIMIT ?",
	allFields,
)

// 默认降序
func (r *Repo) PageGetByUidOrderByMtime(ctx context.Context, bizCode int32, 
	param PageGetByUidOrderByMtimeParam) ([]*Record, error) {
		
	var models []*Record
	handlePageGetByUidOrderByMtimeParam(&param)

	var cond = "mtime DESC, id DESC"
	if param.Order == Asc {
		cond = "mtime ASC, id ASC"
	}
	sql := fmt.Sprintf(sqlPageGetByUidOrderByMtime, cond)

	err := r.db.QueryRowsCtx(ctx, &models, sql, param.Uid, bizCode, ActDo, param.Count)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return models, nil
}

type PageGetByUidOrderByMtimeCursor struct {
	Mtime int64
	Id    int64
}

var sqlPageGetByUidOrderByMtimeWithCursor = fmt.Sprintf(
	"SELECT %s FROM counter_record WHERE uid=? AND biz_code=? AND %%s AND act=? ORDER BY %%s LIMIT ?",
	allFields,
)

// 默认降序
func (r *Repo) PageGetByUidOrderByMtimeWithCursor(ctx context.Context,
	bizCode int32, param PageGetByUidOrderByMtimeParam,
	cursor PageGetByUidOrderByMtimeCursor) ([]*Record, error) {

	handlePageGetByUidOrderByMtimeParam(&param)

	if cursor.Mtime == 0 || cursor.Id == 0 {
		return r.PageGetByUidOrderByMtime(ctx, bizCode, param)
	}

	var (
		cond1 = "(mtime < ? OR (mtime = ? AND id < ?))"
		cond2 = "mtime DESC, id DESC"
	)
	if param.Order == Asc {
		cond1 = "(mtime > ? OR (mtime = ? AND id > ?))"
		cond2 = "mtime ASC, id ASC"
	}
	sql := fmt.Sprintf(sqlPageGetByUidOrderByMtimeWithCursor, cond1, cond2)

	var models []*Record
	err := r.db.QueryRowsCtx(ctx, &models, sql,
		param.Uid, bizCode, cursor.Mtime, cursor.Mtime, cursor.Id, ActDo, param.Count,
	)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return models, nil
}
