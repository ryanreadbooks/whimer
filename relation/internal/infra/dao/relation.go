package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type LinkStatus int8

// Link的状态转移
//
// 初次关注 LinkVacant -> LinkForward/LinkBackward
//
// 单向关注，随后取消关注 LinkForward/LinkBackward -> LinkVacant
//
// 单向关注，随后另一个人互相关注 LinkFowrard/LinkBackward -> LinkMutual
//
// 互相关注，随后其中一个人取消关注 LinkMutual -> LinkForward/LinkBackward
const (
	LinkVacant   LinkStatus = 0  // 没有关系
	LinkForward  LinkStatus = 2  // A单向关注B
	LinkBackward LinkStatus = -2 // B单向关注A
	LinkMutual   LinkStatus = -4 // AB互相关注
)

func (l LinkStatus) IsMutual() bool {
	return l == LinkMutual
}

func (l LinkStatus) IsForward() bool {
	return l == LinkForward
}

func (l LinkStatus) IsBackward() bool {
	return l == LinkBackward
}

func (l LinkStatus) IsVacant() bool {
	return l == LinkVacant
}

type Relation struct {
	// UserAlpha和UserBeta为两个存在关注关系的用户
	//
	// 假设存在a=100,b=200,其实(100,2,200)和(200,-2,100)是相同的一组关系
	// 所以此处需要作出一个规定：插入表中的alpha必须比beta小，从而保证(100,200)和(200,100)不会在数据库产生两条数据
	//
	// 示例：(200,-2,100)==(100,2,200), (200,2,100)==(100,-2,200)
	//			(200,-4,100)==(100,-4,200)
	Id        int64      `db:"id"`
	UserAlpha int64      `db:"alpha"`  // 用户A
	UserBeta  int64      `db:"beta"`   // 用户B
	Link      LinkStatus `db:"link"`   // 用户的关注关系
	Actime    int64      `db:"actime"` // A首次关注B的时间，Unix时间戳
	Bctime    int64      `db:"bctime"` // B首次关注A的时间，Unix时间戳
	Amtime    int64      `db:"amtime"` // A改变对B的关注状态的时间，Unix时间戳
	Bmtime    int64      `db:"bmtime"` // B改变对A的关注状态的时间，Unix时间戳
}

func (r *Relation) IsLinkVacant() bool {
	return r != nil && r.Link == LinkVacant
}

func (r *Relation) IsLinkForward() bool {
	return r != nil && r.Link == LinkForward
}

func (r *Relation) IsLinkBackward() bool {
	return r != nil && r.Link == LinkBackward
}

func (r *Relation) IsLinkMutual() bool {
	return r != nil && r.Link == LinkMutual
}

func (r *Relation) CheckUserAFollowsUserB(userA, userB int64) bool {
	if r == nil {
		return false
	}

	alpha, beta := enforceUidRule(userA, userB)
	if alpha != r.UserAlpha || beta != r.UserBeta {
		return false
	}

	if r.IsLinkMutual() {
		return true
	}

	if alpha == userA {
		return r.IsLinkForward()
	}

	if alpha == userB {
		return r.IsLinkBackward()
	}

	return false
}

type RelationUser struct {
	Id        int64      `db:"id"`
	UserAlpha int64      `db:"alpha"` // 用户A
	UserBeta  int64      `db:"beta"`  // 用户B
	Link      LinkStatus `db:"link"`  // 用户的关注关系
}

// 见[Relation]的uid规则
func enforceRelationRule(r *Relation) *Relation {
	if r.UserAlpha > r.UserBeta {
		// 交换两者
		r.UserAlpha, r.UserBeta = r.UserBeta, r.UserAlpha
		r.Actime, r.Bctime = r.Bctime, r.Actime
		r.Amtime, r.Bmtime = r.Bmtime, r.Amtime
		// 还需要反转关系
		if r.Link != LinkMutual {
			r.Link = -r.Link
		}
	}

	return r
}

func reverseLink(link LinkStatus) LinkStatus {
	if link == LinkVacant || link == LinkMutual {
		return link
	}

	return -link
}

func enforceUidRule(a, b int64) (int64, int64) {
	if a > b {
		return b, a
	}

	return a, b
}

func enforceUidRuleWithLink(a, b int64, link LinkStatus) (int64, int64, LinkStatus) {
	if a > b {
		return b, a, reverseLink(link)
	}

	return a, b, link
}

func newRelationFromAlphaToBeta(a, b int64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkForward,
		Actime:    time.Now().Unix(),
	}
}

func newRelationFromBetaToAlpha(a, b int64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkBackward,
		Bctime:    time.Now().Unix(),
	}
}

func newMutualRelation(a, b int64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkMutual,
	}
}

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

	relationAllFields  = "id,alpha,beta,link,actime,bctime,amtime,bmtime"
	relationUserFields = "id,alpha,beta,link"
	relationFields     = "alpha,beta,link,actime,bctime,amtime,bmtime"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO relation(%s) VALUES(?,?,?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE link=val.link, actime=val.actime, bctime=val.bctime, amtime=val.amtime, bmtime=val.bmtime",
		relationFields)
	sqlUpdateLink      = "UPDATE relation SET link=?, amtime=?, bmtime=? WHERE alpha=? AND beta=?"
	sqlFindByAlphaBeta = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? %%s", relationAllFields)

	unionBaseTemplate = fmt.Sprintf(`
		(SELECT %s FROM relation WHERE id>? AND alpha=? AND link IN (%%d, %%d) LIMIT ?) 
			UNION ALL 
		(SELECT %s FROM relation WHERE id>? AND beta=? AND link IN (%%d, %%d) LIMIT ?) LIMIT ?`,
		relationUserFields, relationUserFields)

	unionBaseTemplateAll = fmt.Sprintf(`
		(SELECT %s FROM relation WHERE alpha=? AND link IN (%%d, %%d)) 
			UNION ALL 
		(SELECT %s FROM relation WHERE beta=? AND link IN (%%d, %%d))`, relationUserFields, relationUserFields)

	sqlUnionTemplate    = strings.ReplaceAll(strings.ReplaceAll(unionBaseTemplate, "\n", ""), "\t", "")
	sqlUnionTemplateAll = strings.ReplaceAll(strings.ReplaceAll(unionBaseTemplateAll, "\n", ""), "\t", "")

	sqlBatchFindUidLinkTo = strings.ReplaceAll(strings.ReplaceAll(fmt.Sprintf(`
		(SELECT %s FROM relation WHERE alpha=? AND beta IN (%%s) AND link IN (%d, %d)) 
			UNION ALL 
		(SELECT %s FROM relation WHERE beta=? AND alpha IN (%%s) AND link IN (%d, %d))`,
		relationUserFields, LinkForward, LinkMutual,
		relationUserFields, LinkBackward, LinkMutual,
	), "\n", ""), "\t", "")

	sqlBatchFindAlphaLinkTo = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta IN (%%s) AND link IN (%d, %d)",
		relationUserFields, LinkForward, LinkMutual)
	sqlBatchFindBetaLinkTo = fmt.Sprintf("SELECT %s FROM relation WHERE beta=? AND alpha IN (%%s) AND link IN (%d, %d)",
		relationUserFields, LinkBackward, LinkMutual)

	// 获取uid关注的人
	sqlFindWhoUidFollows = fmt.Sprintf(sqlUnionTemplate, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取全部uid关注的人
	sqlFindWhoUidFollowsAll = fmt.Sprintf(sqlUnionTemplateAll, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取关注uid的人
	sqlFindWhoFollowsUid = fmt.Sprintf(sqlUnionTemplate, LinkBackward, LinkMutual, LinkForward, LinkMutual)

	sqlFindTemplate       = fmt.Sprintf("SELECT %s FROM relation WHERE %%s=? AND (link IN (%%d,%%d))", relationAllFields)
	sqlFindByAlpha        = fmt.Sprintf(sqlFindTemplate, "alpha", LinkForward, LinkMutual)
	sqlFindByBeta         = fmt.Sprintf(sqlFindTemplate, "beta", LinkBackward, LinkMutual)
	sqlFindAlphaGotLinked = fmt.Sprintf(sqlFindTemplate, "alpha", LinkBackward, LinkMutual)
	sqlFindBetaGotLinked  = fmt.Sprintf(sqlFindTemplate, "beta", LinkForward, LinkMutual)

	// counting
	unionCountTemplate = `
		SELECT SUM(cnt) FROM 
		((SELECT COUNT(*) cnt FROM relation WHERE alpha=? AND link IN (%d, %d)) 
			UNION ALL 
		(SELECT COUNT(*) cnt FROM relation WHERE beta=? AND link IN (%d, %d))) AS total
	`
	sqlUnionCountTemplate = strings.ReplaceAll(strings.ReplaceAll(unionCountTemplate, "\n", ""), "\t", "")
	// 获取uid关注的人的数量
	sqlCountUidFollowings = fmt.Sprintf(sqlUnionCountTemplate, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取关注uid的人的数量
	sqlCountUidFans = fmt.Sprintf(sqlUnionCountTemplate, LinkBackward, LinkMutual, LinkForward, LinkMutual)
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
	uids []int64, next int64, more bool, err error) {

	var (
		rs = make([]*RelationUser, 0, 50)
	)

	uids = []int64{}
	err = d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoUidFollows, offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []int64{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	uids = make([]int64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, r.UserBeta)
		} else {
			uids = append(uids, r.UserAlpha)
		}
	}

	next = rs[len(rs)-1].Id
	if len(rs) == limit {
		more = true
	} else {
		next = 0
	}

	return
}

// 找出uid关注的全部人
func (d *RelationDao) FindAllUidLinkTo(ctx context.Context, uid int64) ([]int64, error) {
	var (
		rs     = make([]*RelationUser, 0, 80)
		others = []int64{}
	)

	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoUidFollowsAll, uid, uid)
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

	others = make([]int64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			others = append(others, r.UserBeta)
		} else {
			others = append(others, r.UserAlpha)
		}
	}

	return others, nil
}

// 批量获取uid和other的关注关系
func (d *RelationDao) BatchFindUidLinkTo(ctx context.Context, uid int64, others []int64) ([]*RelationUser, error) {
	const batchsize = 100

	var relations = make([]*RelationUser, 0, len(others))
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

		var rs = make([]*RelationUser, 0, patchLen)
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
func (d *RelationDao) FindAlphaLinkTo(ctx context.Context, alpha int64) ([]int64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByAlpha, alpha)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []int64{}, nil
		}
		return nil, err
	}

	uids := make([]int64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserBeta)
	}

	return uids, nil
}

// 找到beta关注的人
func (d *RelationDao) FindBetaLinkTo(ctx context.Context, beta int64) ([]int64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByBeta, beta)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []int64{}, nil
		}
		return nil, err
	}

	uids := make([]int64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserAlpha)
	}

	return uids, nil
}

// 找到关注uid的人
func (d *RelationDao) FindUidGotLinked(ctx context.Context, uid int64, offset int64, limit int) (
	uids []int64, next int64, more bool, err error) {
		
	var (
		rs = make([]*RelationUser, 0, 16)
	)

	uids = []int64{}
	err = d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoFollowsUid, offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []int64{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	uids = make([]int64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, r.UserBeta)
		} else {
			uids = append(uids, r.UserAlpha)
		}
	}

	next = rs[len(rs)-1].Id
	if len(rs) == limit {
		more = true
	} else {
		next = 0
	}

	return
}

// 找到关注alpha的人
func (d *RelationDao) FindAlphaGotLinked(ctx context.Context, alpha int64) ([]int64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindAlphaGotLinked, alpha)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []int64{}, nil
		}
		return nil, err
	}

	uids := make([]int64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserBeta)
	}

	return uids, nil
}

// 找到关注beta的人
func (d *RelationDao) FindBetaGotLinked(ctx context.Context, beta int64) ([]int64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindBetaGotLinked, beta)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []int64{}, nil
		}
		return nil, err
	}

	uids := make([]int64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserAlpha)
	}

	return uids, nil
}

// 获取关注uid的人数
func (d *RelationDao) CountUidFans(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidFans, uid, uid)
	return cnt, xsql.ConvertError(err)
}

// 获取uid关注的人数
func (d *RelationDao) CountUidFollowings(ctx context.Context, uid int64) (int64, error) {
	var cnt int64
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidFollowings, uid, uid)
	return cnt, xsql.ConvertError(err)
}
