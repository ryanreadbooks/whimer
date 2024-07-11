package model

import (
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
)

const (
	MaxNicknameLen  = 20
	MaxStyleSignLen = 64
)

// 可以返回给各端的个人信息
type MeInfo struct {
	Uid       uint64 `json:"uid"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar,omitempty"`
	StyleSign string `json:"style_sign"`
	Gender    string `json:"gender"`
}

var (
	GenderMap = map[int8]string{
		0: "未设置",
		1: "男",
		2: "女",
	}
)

func NewMeInfoFromUserBasic(user *userbase.Basic) *MeInfo {
	return &MeInfo{
		Uid:       user.Uid,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		StyleSign: user.StyleSign,
		Gender:    GenderMap[user.Gender],
	}
}
