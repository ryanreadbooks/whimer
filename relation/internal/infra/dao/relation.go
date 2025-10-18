package dao

import (
	"time"
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
	now := time.Now().Unix()
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkForward,
		Actime:    now,
		Amtime:    now,
	}
}

func newRelationFromBetaToAlpha(a, b int64) *Relation {
	now := time.Now().Unix()
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkBackward,
		Bctime:    now,
		Bmtime:    now,
	}
}

func newMutualRelation(a, b int64) *Relation {
	now := time.Now().Unix()
	return &Relation{
		UserAlpha: a,
		UserBeta:  b,
		Link:      LinkMutual,
		Actime:    now,
		Amtime:    now,
		Bctime:    now,
		Bmtime:    now,
	}
}

type UidWithTime struct {
	Uid  int64
	Time int64
}
