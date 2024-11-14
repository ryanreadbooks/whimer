package dao

import (
	"context"
	"errors"
	"fmt"
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

	UserAlpha uint64     `db:"alpha"`  // 用户A
	UserBeta  uint64     `db:"beta"`   // 用户B
	Link      LinkStatus `db:"link"`   // 用户的关注关系
	Actime    int64      `db:"actime"` // A首次关注B的时间，Unix时间戳
	Bctime    int64      `db:"bctime"` // B首次关注A的时间，Unix时间戳
	Amtime    int64      `db:"amtime"` // A改变对B的关注状态的时间，Unix时间戳
	Bmtime    int64      `db:"bmtime"` // B改变对A的关注状态的时间，Unix时间戳
}

// 见[Relation]的uid规则
func enforceUserRule(r *Relation) *Relation {
	if r.UserAlpha > r.UserBeta {
		// 交换两者
		r.UserAlpha, r.UserBeta = r.UserBeta, r.UserAlpha
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
	relationFields = "alpha,beta,link,actime,bctime,amtime,bmtime"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO relation(%s) VALUES(?,?,?,?,?,?,?) AS val "+
		"ON DUPLICATE KEY UPDATE link=val.link, amtime=val.amtime, bmtime=val.bmtime", relationFields)
	sqlUpdateLink                   = "UPDATE relation SET link=?, amtime=?, bmtime=? WHERE alpha=? AND beta=?"
	sqlFindByAlphaBeta              = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=?", relationFields)
	sqlFindByAlphaBetaLink          = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? AND link=?", relationFields)
	sqlFindByAlphaBetaLinkForUpdate = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? AND link=? FOR UPDATE", relationFields)

	sqlFindFollowTemplate = fmt.Sprintf("SELECT %s FROM relation WHERE (alpha=? AND (link=%%d OR link=%%d)) OR (beta=? AND (link=%%d OR link=%%d))", relationFields)
	// 获取uid关注的人
	sqlFindWhoUidFollows = fmt.Sprintf(sqlFindFollowTemplate, LinkForward, LinkMutual, LinkBackward, LinkMutual)
	// 获取关注uid的人
	sqlFindWhoFollowUid = fmt.Sprintf(sqlFindFollowTemplate, LinkBackward, LinkMutual, LinkForward, LinkMutual)

	sqlFindTemplate = fmt.Sprintf("SELECT %s FROM relation WHERE %%s=? AND (link=%%d OR link=%%d)", relationFields)

	sqlFindByAlpha        = fmt.Sprintf(sqlFindTemplate, "alpha", LinkForward, LinkMutual)
	sqlFindByBeta         = fmt.Sprintf(sqlFindTemplate, "beta", LinkBackward, LinkMutual)
	sqlFindAlphaGotLinked = fmt.Sprintf(sqlFindTemplate, "alpha", LinkBackward, LinkMutual)
	sqlFindBetaGotLinked  = fmt.Sprintf(sqlFindTemplate, "beta", LinkForward, LinkMutual)
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
		r.Bctime,
	)

	return xsql.ConvertError(err)
}

func (d *RelationDao) UpdateLink(ctx context.Context, r *Relation) error {
	r = enforceUserRule(r)
	_, err := d.db.ExecCtx(ctx, sqlUpdateLink, r.Link, r.Amtime, r.Bmtime, r.UserAlpha, r.UserBeta)
	return xsql.ConvertError(err)
}

func (d *RelationDao) FindByAlphaBeta(ctx context.Context, a, b uint64) (*Relation, error) {
	a, b = enforceUidRule(a, b)
	var r Relation
	err := d.db.QueryRowCtx(ctx, &r, sqlFindByAlphaBeta, a, b)
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

// 找到uid关注的人
// alpha=uid and link=Forward/Mutual or beta=uid and link=Backward/Mutual
// 找到发出关注连接的用户存在的用户关系
func (d *RelationDao) FindUidLinkTo(ctx context.Context, uid uint64) ([]uint64, error) {
	var (
		rs = make([]*Relation, 0, 16)
	)

	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoUidFollows, uid, uid)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, r.UserBeta)
		} else {
			uids = append(uids, r.UserAlpha)
		}
	}

	return uids, nil
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
func (d *RelationDao) FindUidGotLinked(ctx context.Context, uid uint64) ([]uint64, error) {
	var (
		rs = make([]*Relation, 0, 16)
	)

	err := d.db.QueryRowsCtx(ctx, &rs, sqlFindWhoFollowUid, uid, uid)
	if err != nil {
		err = xsql.ConvertError(err)
		if errors.Is(err, xsql.ErrNoRecord) {
			return []uint64{}, nil
		}
		return nil, err
	}

	uids := make([]uint64, 0, len(rs))
	for _, r := range rs {
		if r.UserAlpha == uid {
			uids = append(uids, r.UserBeta)
		} else {
			uids = append(uids, r.UserAlpha)
		}
	}

	return uids, nil
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
