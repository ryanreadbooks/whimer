package vo

import noteentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"

// hover卡片基础信息
type HoverBasicInfo struct {
	Nickname  string
	StyleSign string
	Avatar    string
}

// hover卡片交互信息
type HoverInteraction struct {
	Fans    string
	Follows string
}

// hover卡片关系信息
type HoverRelation struct {
	Status RelationStatus
}

// 用户卡片信息
type HoverInfo struct {
	BasicInfo   HoverBasicInfo
	Interaction HoverInteraction
	Relation    HoverRelation
	RecentPosts []*noteentity.RecentPost
}
