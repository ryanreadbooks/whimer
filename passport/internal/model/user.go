package model

import (
	"unicode/utf8"

	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dao"
)

const (
	MaxNicknameLen  = 20
	MaxStyleSignLen = 64
	MaxAvatarBytes  = 500 * 1024
)

// 可以返回给各端的个人信息
type UserInfo struct {
	Uid       int64  `json:"uid"`
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

	GenderMapConvert = map[string]int8{
		"保密": 0,
		"男":  1,
		"女":  2,
	}
)

func NewUserInfoFromUserBase(user *dao.UserBase) *UserInfo {
	return &UserInfo{
		Uid:       user.Uid,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		StyleSign: user.StyleSign,
		Gender:    GenderMap[user.Gender],
	}
}

func (u *UserInfo) ToUserBase() *dao.UserBase {
	return &dao.UserBase{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		StyleSign: u.StyleSign,
		Gender:    GenderMapConvert[u.Gender],
	}
}

func (u *UserInfo) ToPb() *userv1.UserInfo {
	return &userv1.UserInfo{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		StyleSign: u.StyleSign,
		Avatar:    u.Avatar,
		Gender:    u.Gender,
	}
}

type UpdateUserRequest struct {
	Uid       int64  `json:"uid"`
	Nickname  string `json:"nickname"`
	StyleSign string `json:"style_sign"`
	Gender    int8   `json:"gender"`
}

func (r *UpdateUserRequest) Validate() error {
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

