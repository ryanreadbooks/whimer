package dao

import (
	"context"
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

func NewRelationFromAlphaToBeta(a, b uint64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkForward,
		Actime:    time.Now().Unix(),
	}
}

func NewRelationFromBetaToAlpha(a, b uint64) *Relation {
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkBackward,
		Bctime:    time.Now().Unix(),
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
	fields = "alpah,beta,link,actime,bctime,amtime,bmtime"
)

var (
	sqlInsert = fmt.Sprintf("INSERT INTO relation(%s) VALUES(?,?,?,?,?,?,?) AS val"+
		"ON DUPLICATE KEY UPDATE link=val.link, amtime=val.amtime, bmtime=val.bmtime", fields)
	sqlUpdateLink        = "UPDATE relation SET link=?, amtime=?, bmtime=? WHERE alpha=? AND beta=?"
	sqlFindByUids        = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=?", fields)
	sqlFindByUidsAndLink = fmt.Sprintf("SELECT %s FROM relation WHERE alpha=? AND beta=? AND link=?", fields)
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

func (d *RelationDao) FindByUids(ctx context.Context, a, b uint64) (*Relation, error) {
	a, b = enforceUidRule(a, b)
	var r Relation
	err := d.db.QueryRowCtx(ctx, &r, sqlFindByUids, a, b)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &r, nil
}

func (d *RelationDao) FindByUidsAndLink(ctx context.Context, a, b uint64, link LinkStatus) (*Relation, error) {
	a, b = enforceUidRule(a, b)
	var r Relation
	err := d.db.QueryRowCtx(ctx, &r, sqlFindByUidsAndLink, a, b, link)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &r, nil
}

// 找到uid关注的人
// alpha=uid and link=Forward/Mutual or beta=uid and link=Backward/Mutual
// 找到发出关注连接的用户的用户关系
func (d *Relation) FindLinkedUid(ctx context.Context, uid uint64) ([]*Relation, error) {
	var rs = make([]*Relation, 0, 16)

	return rs, nil
}
