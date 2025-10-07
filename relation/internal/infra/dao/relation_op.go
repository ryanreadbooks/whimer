package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type RelationDao struct {
	db *xsql.DB
}

func NewRelationDao(db *xsql.DB) *RelationDao {
	return &RelationDao{
		db: db,
	}
}

// all sqls here
const (
	sqlForUpdate = "FOR UPDATE"

	relationFields     = "id,alpha,beta,link,actime,bctime,amtime,bmtime"
	relationFieldsNoId = "alpha,beta,link,actime,bctime,amtime,bmtime"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO relation(%s) VALUES(?,?,?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE link=val.link, actime=val.actime, bctime=val.bctime, amtime=val.amtime, bmtime=val.bmtime",
		relationFieldsNoId)
	sqlUpdateLink      = "UPDATE relation SET link=?, amtime=?, bmtime=? WHERE alpha=? AND beta=?"
	sqlFindByAlphaBeta = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? %%s", relationFields)

	sqlUnionTemplate = fmt.Sprintf(""+
		"(SELECT %s FROM relation WHERE id>? AND alpha=? AND link IN (%%d, %%d) LIMIT ?) "+
		"UNION ALL "+
		"(SELECT %s FROM relation WHERE id>? AND beta=? AND link IN (%%d, %%d) LIMIT ?) ORDER BY id LIMIT ?",
		relationFields, relationFields)

	sqlUnionTemplateAll = fmt.Sprintf(""+
		"(SELECT %s FROM relation WHERE alpha=? AND link IN (%%d, %%d)) "+
		"UNION ALL "+
		"(SELECT %s FROM relation WHERE beta=? AND link IN (%%d, %%d)) ORDER BY id",
		relationFields, relationFields)

	sqlBatchFindUidLinkTo = fmt.Sprintf(""+
		"(SELECT %s FROM relation WHERE alpha=? AND beta IN (%%s) AND link IN (%d, %d)) "+
		"UNION ALL "+
		"(SELECT %s FROM relation WHERE beta=? AND alpha IN (%%s) AND link IN (%d, %d)) ORDER BY id",
		relationFields, LinkForward, LinkMutual,
		relationFields, LinkBackward, LinkMutual)

	sqlBatchFindAlphaLinkTo = fmt.Sprintf(
		"SELECT %s FROM relation WHERE alpha=? AND beta IN (%%s) AND link IN (%d, %d)",
		relationFields, LinkForward, LinkMutual)
	sqlBatchFindBetaLinkTo = fmt.Sprintf(
		"SELECT %s FROM relation WHERE beta=? AND alpha IN (%%s) AND link IN (%d, %d)",
		relationFields, LinkBackward, LinkMutual)

	// 获取uid关注的人
	sqlFindUidLinkTo = fmt.Sprintf(sqlUnionTemplate, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取全部uid关注的人
	sqlFindUidLinkTo2 = fmt.Sprintf(sqlUnionTemplateAll, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取关注uid的人
	sqlFindWhoFollowsUid = fmt.Sprintf(sqlUnionTemplate, LinkBackward, LinkMutual, LinkForward, LinkMutual)

	sqlFindTemplate = fmt.Sprintf(
		"SELECT %s FROM relation WHERE %%s=? AND (link IN (%%d,%%d))", relationFields)
	sqlFindByAlpha        = fmt.Sprintf(sqlFindTemplate, "alpha", LinkForward, LinkMutual)
	sqlFindByBeta         = fmt.Sprintf(sqlFindTemplate, "beta", LinkBackward, LinkMutual)
	sqlFindAlphaGotLinked = fmt.Sprintf(sqlFindTemplate, "alpha", LinkBackward, LinkMutual)
	sqlFindBetaGotLinked  = fmt.Sprintf(sqlFindTemplate, "beta", LinkForward, LinkMutual)

	// counting
	sqlUnionCountTemplate = "" +
		"SELECT SUM(cnt) FROM " +
		"((SELECT COUNT(*) cnt FROM relation WHERE alpha=? AND link IN (%d, %d)) " +
		"UNION ALL " +
		"(SELECT COUNT(*) cnt FROM relation WHERE beta=? AND link IN (%d, %d))) AS total"

	// 获取关注uid的人的数量
	sqlCountUidGotLinked = fmt.Sprintf(sqlUnionCountTemplate, LinkBackward, LinkMutual, LinkForward, LinkMutual)

	// 获取uid关注的人的数量
	sqlCountUidLinkTo = fmt.Sprintf(sqlUnionCountTemplate, LinkForward, LinkMutual, LinkBackward, LinkMutual)

	sqlPageGetLinksTemplate = "" +
		"WITH combined AS (" +
		"SELECT id,alpha,beta,link,actime AS ctime, amtime AS mtime FROM relation WHERE alpha=? AND link IN (%d, %d) " +
		"UNION ALL " +
		"SELECT id,alpha,beta,link,bctime AS ctime, bmtime AS mtime FROM relation WHERE beta=? AND link IN (%d, %d)" +
		") " +
		"SELECT * FROM combined ORDER BY mtime DESC LIMIT ?,?"

	// 分页获取关注uid的人
	sqlPageGetUidGotLinked = fmt.Sprintf(sqlPageGetLinksTemplate,
		LinkBackward, LinkMutual,
		LinkForward, LinkMutual,
	)

	// 分页获取uid关注的人
	sqlPageGetUidLinkTo = fmt.Sprintf(sqlPageGetLinksTemplate,
		LinkForward, LinkMutual,
		LinkBackward, LinkMutual,
	)
)

// 插入/更新一条记录
func (d *RelationDao) Insert(ctx context.Context, r *Relation) error {
	r = enforceRelationRule(r)
	_, err := d.db.ExecCtx(ctx, sqlInsert,
		r.UserAlpha,
		r.UserBeta,
		r.Link,
		r.Actime,
		r.Bctime,
		r.Amtime,
		r.Bmtime,
	)

	return xsql.ConvertError(err)
}

// batchInsert 批量插入/更新多条记录
func (d *RelationDao) batchInsert(ctx context.Context, relations []*Relation) error {
	if len(relations) == 0 {
		return nil
	}

	for i := range relations {
		relations[i] = enforceRelationRule(relations[i])
	}

	var args []any
	var valuePlaceholders []string

	for _, r := range relations {
		args = append(args, r.UserAlpha, r.UserBeta, r.Link, r.Actime, r.Bctime, r.Amtime, r.Bmtime)
		valuePlaceholders = append(valuePlaceholders, "(?,?,?,?,?,?,?)")
	}

	sqlBatchInsert := fmt.Sprintf(""+
		"INSERT INTO relation(%s) VALUES %s ON DUPLICATE KEY UPDATE link=VALUES(link), "+
		"actime=VALUES(actime), bctime=VALUES(bctime), amtime=VALUES(amtime), bmtime=VALUES(bmtime)",
		relationFieldsNoId,
		strings.Join(valuePlaceholders, ","),
	)

	_, err := d.db.ExecCtx(ctx, sqlBatchInsert, args...)

	return xsql.ConvertError(err)
}

func (d *RelationDao) UpdateLink(ctx context.Context, r *Relation) error {
	r = enforceRelationRule(r)
	_, err := d.db.ExecCtx(ctx, sqlUpdateLink, r.Link, r.Amtime, r.Bmtime, r.UserAlpha, r.UserBeta)
	return xsql.ConvertError(err)
}

func (d *RelationDao) FindByAlphaBeta(ctx context.Context, a, b int64, forUpdate bool) (*Relation, error) {
	a, b = enforceUidRule(a, b)
	var r Relation
	sql := fmt.Sprintf(sqlFindByAlphaBeta, "")
	if forUpdate {
		sql = fmt.Sprintf(sqlFindByAlphaBeta, sqlForUpdate)
	}

	err := d.db.QueryRowCtx(ctx, &r, sql, a, b)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &r, nil
}

// 找到uid关注的人 (找到发出关注连接的用户存在的用户关系)
//
//	alpha=uid and link=Forward/Mutual or beta=uid and link=Backward/Mutual
func (d *RelationDao) FindUidLinkTo(ctx context.Context, uid int64, offset int64, limit int) (
	uids []UidWithTime, next int64, more bool, err error) {

	var (
		rs = make([]*Relation, 0, limit)
	)

	uids = []UidWithTime{}
	limit += 1 // 多查一条
	err = d.db.QueryRowsCtx(ctx, &rs, sqlFindUidLinkTo, offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []UidWithTime{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	rsLen := len(rs)
	if rsLen == limit {
		rs = rs[:rsLen-1]
		more = true
		next = rs[len(rs)-1].Id
	} else {
		more = false
		next = 0
	}

	uids = make([]UidWithTime, 0, rsLen)
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, UidWithTime{r.UserBeta, r.Amtime})
		} else {
			uids = append(uids, UidWithTime{r.UserAlpha, r.Bmtime})
		}
	}

	return
}

// 找出uid关注的全部人
func (d *RelationDao) FindAllUidLinkTo(ctx context.Context, uid int64) ([]UidWithTime, error) {
	var (
		rs     = make([]*Relation, 0, 80)
		others = []UidWithTime{}
	)

	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindUidLinkTo2, uid, uid)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return others, nil
		}

		return others, err
	}

	if len(rs) == 0 {
		return others, nil
	}

	others = make([]UidWithTime, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			others = append(others, UidWithTime{r.UserBeta, r.Amtime})
		} else {
			others = append(others, UidWithTime{r.UserAlpha, r.Bmtime})
		}
	}

	return others, nil
}

// 批量获取uid和other的关注关系
func (d *RelationDao) BatchFindUidLinkTo(ctx context.Context, uid int64, others []int64) ([]*Relation, error) {
	const batchsize = 100

	var relations = make([]*Relation, 0, len(others))
	err := xslice.BatchExec(others, batchsize, func(start, end int) error {
		patch := others[start:end]
		patchLen := len(patch)
		lesser := make([]int64, 0, patchLen)
		greater := make([]int64, 0, patchLen)
		for _, ou := range patch {
			if ou > uid {
				greater = append(greater, ou)
			} else if ou < uid {
				lesser = append(lesser, ou)
			}
		}

		var (
			sql  string
			args []any = make([]any, 0, 3)
		)
		if len(lesser) != 0 && len(greater) != 0 {
			sql = fmt.Sprintf(sqlBatchFindUidLinkTo,
				xslice.JoinInts(greater),
				xslice.JoinInts(lesser),
			)
			args = append(args, uid, uid)
		} else if len(lesser) != 0 && len(greater) == 0 {
			sql = fmt.Sprintf(sqlBatchFindBetaLinkTo, xslice.JoinInts(lesser))
			args = append(args, uid)
		} else if len(lesser) == 0 && len(greater) != 0 {
			sql = fmt.Sprintf(sqlBatchFindAlphaLinkTo, xslice.JoinInts(greater))
			args = append(args, uid)
		} else {
			// lesser == 0 and greater == 0
			return nil
		}

		var rs = make([]*Relation, 0, patchLen)
		err := d.db.QueryRowsCtx(ctx, &rs, sql, args...)
		if err != nil {
			return xsql.ConvertError(err)
		}

		relations = append(relations, rs...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return relations, nil
}

// 找到alpha关注的人
func (d *RelationDao) FindAlphaLinkTo(ctx context.Context, alpha int64) ([]UidWithTime, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByAlpha, alpha)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []UidWithTime{}, nil
		}
		return nil, err
	}

	uids := make([]UidWithTime, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, UidWithTime{r.UserBeta, r.Amtime})
	}

	return uids, nil
}

// 找到beta关注的人
func (d *RelationDao) FindBetaLinkTo(ctx context.Context, beta int64) ([]UidWithTime, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByBeta, beta)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []UidWithTime{}, nil
		}
		return nil, err
	}

	uids := make([]UidWithTime, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, UidWithTime{r.UserAlpha, r.Bmtime})
	}

	return uids, nil
}

// 找到关注uid的人
func (d *RelationDao) FindUidGotLinked(ctx context.Context, uid int64, offset int64, limit int) (
	uids []UidWithTime, next int64, more bool, err error) {

	var (
		rs = make([]*Relation, 0, limit)
	)

	uids = []UidWithTime{}
	limit += 1 // 多查一条
	err = d.db.QueryRowsCtx(ctx, &rs,
		sqlFindWhoFollowsUid,
		offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []UidWithTime{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	rsLen := len(rs)
	if rsLen == limit {
		rs = rs[:rsLen-1]
		more = true
		next = rs[rsLen-1].Id
	} else {
		more = false
		next = 0
	}

	uids = make([]UidWithTime, 0, rsLen)
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, UidWithTime{Uid: r.UserBeta, Time: r.Bmtime}) // user_beta在bmtime关注了user_alpha
		} else {
			uids = append(uids, UidWithTime{Uid: r.UserAlpha, Time: r.Amtime}) // user_alpha在amtime关注了user_beta
		}
	}

	return
}

// 找到关注alpha的人
func (d *RelationDao) FindAlphaGotLinked(ctx context.Context, alpha int64) ([]UidWithTime, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindAlphaGotLinked, alpha)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []UidWithTime{}, nil
		}
		return nil, err
	}

	uids := make([]UidWithTime, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, UidWithTime{Uid: r.UserBeta, Time: r.Bmtime})
	}

	return uids, nil
}

// 找到关注beta的人
func (d *RelationDao) FindBetaGotLinked(ctx context.Context, beta int64) ([]UidWithTime, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindBetaGotLinked, beta)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []UidWithTime{}, nil
		}
		return nil, err
	}

	uids := make([]UidWithTime, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, UidWithTime{Uid: r.UserAlpha, Time: r.Amtime})
	}

	return uids, nil
}

// 获取关注uid的人数
func (d *RelationDao) CountUidGotLinked(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidGotLinked, uid, uid)
	return cnt, xsql.ConvertError(err)
}

// 获取uid关注的人数
func (d *RelationDao) CountUidLinkTo(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidLinkTo, uid, uid)
	return cnt, xsql.ConvertError(err)
}

type PageGetRelationResult struct {
	Id    int64      `db:"id"`
	Alpha int64      `db:"alpha"`
	Beta  int64      `db:"beta"`
	Link  LinkStatus `db:"link"`
	Ctime int64      `db:"ctime"`
	Mtime int64      `db:"mtime"`
}

// 分页获取关注uid的人
func (d *RelationDao) PageGetUidGotLinked(ctx context.Context, uid int64, page, count int32) ([]*PageGetRelationResult, error) {
	var res = make([]*PageGetRelationResult, 0, count)
	limit := (page - 1) * count
	err := d.db.QueryRowsCtx(ctx, &res, sqlPageGetUidGotLinked, uid, uid, limit, count)
	return res, xsql.ConvertError(err)
}

// 分页获取uid关注的人
func (d *RelationDao) PageGetUidLinkTo(ctx context.Context, uid int64, page, count int32) ([]*PageGetRelationResult, error) {
	var res = make([]*PageGetRelationResult, 0, count)
	limit := (page - 1) * count
	err := d.db.QueryRowsCtx(ctx, &res, sqlPageGetUidLinkTo, uid, uid, limit, count)
	return res, xsql.ConvertError(err)
}
