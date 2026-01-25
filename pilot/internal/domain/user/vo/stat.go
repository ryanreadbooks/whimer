package vo

// 用户统计信息
type UserStat struct {
	Posted     int64 // 投稿数
	Fans       int64 // 粉丝数
	Followings int64 // 关注数
}

// 关系状态
type RelationStatus string

const (
	RelationFollowing RelationStatus = "following"
	RelationNone      RelationStatus = "none"
)

// 用户信息（附带关注状态）
type UserWithRelation struct {
	User     *User
	Relation RelationStatus
}
