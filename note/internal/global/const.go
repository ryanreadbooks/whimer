package global

const (
	PrivacyPublic  = 1
	PrivacyPrivate = 2
)

const (
	AssetTypeImage        = 1 // 完整图片
	AssetTypeVideo        = 2 // TODO 视频
	AssetTypeImagePreview = 3 // 低质量图片 一般用作封面
)

// 计数服务的业务码
const (
	NoteLikeBizcode int32 = 20001 + iota // 点赞
)
