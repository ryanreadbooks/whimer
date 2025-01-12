package dao

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type LinkStatus int8

// Link的状态转移
//
// 初次关注 LinkVacant -> LinkForward/LinkBackward
// 单向关注，随后取消关注 LinkForward/LinkBackward -> LinkVacant
// 单向关注，随后另一个人互相关注 LinkFowrard/LinkBackward -> LinkMutual
// 互相关注，随后其中一个人取消关注 LinkMutual -> LinkForward/LinkBackward
const (
	LinkVacant   LinkStatus = 0  // 没有关系
	LinkForward  LinkStatus = 2  // A单向关注B
	LinkBackward LinkStatus = -2 // B单向关注A
	LinkMutual   LinkStatus = -4 // AB互相关注
)

type Relation struct {
	// UserAlpha和UserBeta为两个存在关注关系的用户
	//
	// 假设存在a=100,b=200,其实(100,2,200)和(200,-2,100)是相同的一组关系
	// 所以此处需要作出一个规定：插入表中的alpha必须比beta小，从而保证(100,200)和(200,100)不会在数据库产生两条数据
	//
	// 示例：(200,-2,100)==(100,2,200), (200,2,100)==(100,-2,200)
	//			(200,-4,100)==(100,-4,200)
	Id        uint64     `db:"id"`
	UserAlpha uint64     `db:"alpha"`  // 用户A
	UserBeta  uint64     `db:"beta"`   // 用户B
	Link      LinkStatus `db:"link"`   // 用户的关注关系
	Actime    int64      `db:"actime"` // A首次关注B的时间，Unix时间戳
	Bctime    int64      `db:"bctime"` // B首次关注A的时间，Unix时间戳
	Amtime    int64      `db:"amtime"` // A改变对B的关注状态的时间，Unix时间戳
	Bmtime    int64      `db:"bmtime"` // B改变对A的关注状态的时间，Unix时间戳
}

type RelationUser struct {
	Id        uint64     `db:"id"`
	UserAlpha uint64     `db:"alpha"` // 用户A
	UserBeta  uint64     `db:"beta"`  // 用户B
	Link      LinkStatus `db:"link"`  // 用户的关注关系
}

// 见[Relation]的uid规则
func enforceUserRule(r *Relation) *Relation {
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

func enforceUidRule(a, b uint64) (uint64, uint64) {
	if a > b {
		return b, a
	}

	return a, b
}

func newRelationFromAlphaToBeta(a, b uint64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkForward,
		Actime:    time.Now().Unix(),
	}
}

func newRelationFromBetaToAlpha(a, b uint64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkBackward,
		Bctime:    time.Now().Unix(),
	}
}

func newMutualRelation(a, b uint64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkMutual,
	}
}

type RelationDao struct {
	db    *xsql.DB
	cache *redis.Redis
}

func NewRelationDao(db *xsql.DB, c *redis.Redis) *RelationDao {
	return &RelationDao{
		db:    db,
		cache: c,
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
		"ON DUPLICATE KEY UPDATE link=val.link, actime=val.actime, bctime=val.bctime, amtime=val.amtime, bmtime=val.bmtime", relationFields)
	sqlUpdateLink                   = "UPDATE relation SET link=?, amtime=?, bmtime=? WHERE alpha=? AND beta=?"
	sqlFindByAlphaBeta              = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? %%s", relationAllFields)
	sqlFindByAlphaBetaLink          = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? AND link=?", relationAllFields)
	sqlFindByAlphaBetaLinkForUpdate = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? AND link=? FOR UPDATE", relationAllFields)

	unionBaseTemplate = fmt.Sprintf(`
		(SELECT %s FROM relation WHERE id>? AND alpha=? AND link IN (%%d, %%d) LIMIT ?) 
			UNION ALL 
		(SELECT %s FROM relation WHERE id>? AND beta=? AND link IN (%%d, %%d) LIMIT ?) LIMIT ?`, relationUserFields, relationUserFields)

	unionBaseTemplateAll = fmt.Sprintf(`
		(SELECT %s FROM relation WHERE alpha=? AND link IN (%%d, %%d)) 
			UNION ALL 
		(SELECT %s FROM relation WHERE beta=? AND link IN (%%d, %%d))`, relationUserFields, relationUserFields)

	sqlUnionTemplate    = strings.ReplaceAll(strings.ReplaceAll(unionBaseTemplate, "\n", ""), "\t", "")
	sqlUnionTemplateAll = strings.ReplaceAll(strings.ReplaceAll(unionBaseTemplateAll, "\n", ""), "\t", "")

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
	// 检查是否关注了某人
	sqlFindUidRelation = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=?", relationAllFields)
)

// 插入/更新一条记录
func (d *RelationDao) Insert(ctx context.Context, r *Relation) error {
	r = enforceUserRule(r)
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
	r = enforceUserRule(r)
	_, err := d.db.ExecCtx(ctx, sqlUpdateLink, r.Link, r.Amtime, r.Bmtime, r.UserAlpha, r.UserBeta)
	return xsql.ConvertError(err)
}

func (d *RelationDao) FindByAlphaBeta(ctx context.Context, a, b uint64, forUpdate bool) (*Relation, error) {
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

func (d *RelationDao) FindByAlphaBetaAndLink(ctx context.Context, a, b uint64, link LinkStatus, forUpdate bool) (*Relation, error) {
	a, b = enforceUidRule(a, b)
	var (
		r   Relation
		err error
	)
	if !forUpdate {
		err = d.db.QueryRowCtx(ctx, &r, sqlFindByAlphaBetaLink, a, b, link)
	} else {
		err = d.db.QueryRowCtx(ctx, &r, sqlFindByAlphaBetaLinkForUpdate, a, b, link)
	}
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &r, nil
}

// 找到uid关注的人 (找到发出关注连接的用户存在的用户关系)
//
//	alpha=uid and link=Forward/Mutual or beta=uid and link=Backward/Mutual
func (d *RelationDao) FindUidLinkTo(ctx context.Context, uid, offset uint64, limit int) (uids []uint64, next uint64, more bool, err error) {
	var (
		rs = make([]*RelationUser, 0, 50)
	)

	uids = []uint64{}
	err = d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoUidFollows, offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []uint64{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	uids = make([]uint64, 0, len(rs))
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
func (d *RelationDao) FindAllUidLinkTo(ctx context.Context, uid uint64) ([]uint64, error) {
	var (
		rs     = make([]*RelationUser, 0, 80)
		others = []uint64{}
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

	others = make([]uint64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			others = append(others, r.UserBeta)
		} else {
			others = append(others, r.UserAlpha)
		}
	}

	return others, nil
}

// 找到alpha关注的人
func (d *RelationDao) FindAlphaLinkTo(ctx context.Context, alpha uint64) ([]uint64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByAlpha, alpha)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserBeta)
	}

	return uids, nil
}

// 找到beta关注的人
func (d *RelationDao) FindBetaLinkTo(ctx context.Context, beta uint64) ([]uint64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindByBeta, beta)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserAlpha)
	}

	return uids, nil
}

// 找到关注uid的人
func (d *RelationDao) FindUidGotLinked(ctx context.Context, uid, offset uint64, limit int) (uids []uint64, next uint64, more bool, err error) {
	var (
		rs = make([]*RelationUser, 0, 16)
	)

	uids = []uint64{}
	err = d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoFollowsUid, offset, uid, limit, offset, uid, limit, limit)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			err = nil
			uids = []uint64{}
			return
		}
		return
	}

	if len(rs) == 0 {
		return
	}

	uids = make([]uint64, 0, len(rs))
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
func (d *RelationDao) FindAlphaGotLinked(ctx context.Context, alpha uint64) ([]uint64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindAlphaGotLinked, alpha)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserBeta)
	}

	return uids, nil
}

// 找到关注beta的人
func (d *RelationDao) FindBetaGotLinked(ctx context.Context, beta uint64) ([]uint64, error) {
	var rs = make([]*Relation, 0, 16)
	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindBetaGotLinked, beta)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		uids = append(uids, r.UserAlpha)
	}

	return uids, nil
}

// 获取关注uid的人数
func (d *RelationDao) CountUidFans(ctx context.Context, uid uint64) (uint64, error) {
	var cnt uint64
	// TODO use cache
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidFans, uid, uid)
	return cnt, xsql.ConvertError(err)
}

// 获取uid关注的人数
func (d *RelationDao) CountUidFollowings(ctx context.Context, uid uint64) (uint64, error) {
	var cnt uint64
	// TODO use cache
	err := d.db.QueryRowCtx(ctx, &cnt, sqlCountUidFollowings, uid, uid)
	return cnt, xsql.ConvertError(err)
}
