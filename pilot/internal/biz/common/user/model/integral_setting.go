package model

type BitMapSettingMask int64

// 位图表示用户设置
const (
	ShowNoteLikesSettingMask BitMapSettingMask = 1 << iota // 公开点赞记录
)

func SetFlagsBit(flags int64, mask BitMapSettingMask) int64 {
	return flags | int64(mask)
}

func UnsetFlagsBit(flags int64, mask BitMapSettingMask) int64 {
	return flags & ^int64(mask)
}

func UpdateFlagsBit(flags int64, mask BitMapSettingMask, set bool) int64 {
	if set {
		return SetFlagsBit(flags, mask)
	}

	return UnsetFlagsBit(flags, mask)
}

func ShouldShowNoteLikes(flags int64) bool {
	return (flags & int64(ShowNoteLikesSettingMask)) != 0
}
