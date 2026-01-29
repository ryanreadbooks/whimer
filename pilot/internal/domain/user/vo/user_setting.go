package vo

// 关系服务设置
type RelationSetting struct {
	ShowFanList    bool // 是否展示粉丝列表
	ShowFollowList bool // 是否展示关注列表
}

// 完整的用户设置（聚合本地+远程）
type FullUserSetting struct {
	// 本地设置
	ShowNoteLikes bool

	// 关系服务设置
	ShowFanList    bool
	ShowFollowList bool
}
