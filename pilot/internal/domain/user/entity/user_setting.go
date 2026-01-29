package entity

// UserSetting 用户设置领域实体
type UserSetting struct {
	Uid   int64
	Flags int64
	Ext   []byte
	Ctime int64
	Utime int64
}

// 位图掩码
type SettingFlagMask int64

const (
	ShowNoteLikesMask SettingFlagMask = 1 << iota // 公开点赞记录
)

// ShouldShowNoteLikes 是否展示点赞记录
func (s *UserSetting) ShouldShowNoteLikes() bool {
	return (s.Flags & int64(ShowNoteLikesMask)) != 0
}

// SetShowNoteLikes 设置是否展示点赞记录
func (s *UserSetting) SetShowNoteLikes(show bool) {
	if show {
		s.Flags |= int64(ShowNoteLikesMask)
	} else {
		s.Flags &= ^int64(ShowNoteLikesMask)
	}
}
