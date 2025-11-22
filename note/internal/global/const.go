package global

import v1 "github.com/ryanreadbooks/whimer/note/api/v1"

const (
	PrivacyPublic  = int8(v1.NotePrivacy_PUBLIC)
	PrivacyPrivate = int8(v1.NotePrivacy_PRIVATE)
)

const (
	AssetTypeImage = int8(v1.NoteAssetType_IMAGE) // 完整图片
	AssetTypeVideo = int8(v1.NoteAssetType_VIDEO) // TODO 视频
)

// 计数服务的业务码
const (
	NoteLikeBizcode int32 = 20001 + iota // 点赞
)
