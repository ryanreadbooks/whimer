package profile

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
)

const (
	MaxNicknameLen  = 20
	MaxStyleSignLen = 64
	MaxAvatarBytes  = 500 * 1024
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
		0: "保密",
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

type UpdateMeReq struct {
	Uid       uint64 `json:"uid"`
	Nickname  string `json:"nickname"`
	StyleSign string `json:"style_sign"`
	Gender    int8   `json:"gender"`
}

func (r *UpdateMeReq) Validate() error {
	if r == nil {
		return global.ErrArgs
	}

	if r.Uid <= 0 {
		return global.ErrInvalidUid
	}

	nickLen := utf8.RuneCountInString(r.Nickname)
	if nickLen > MaxNicknameLen {
		return global.ErrNickNameTooLong
	}

	if nickLen <= 0 {
		return global.ErrNicknameTooShort
	}

	if utf8.RuneCountInString(r.StyleSign) > MaxStyleSignLen {
		return global.ErrStyleSignTooLong
	}

	if _, ok := GenderMap[r.Gender]; !ok {
		return global.ErrInvalidGender
	}

	return nil
}

// 上传头像响应结果
type UploadAvatarRes struct {
	Url string `json:"avatar_url"`
}
