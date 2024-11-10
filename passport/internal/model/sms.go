package model

import (
	"regexp"

	global "github.com/ryanreadbooks/whimer/passport/internal/global"
)

var (
	telRegx  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	codeRegx = regexp.MustCompile(`^[1-9][0-9]{5}$`)
)

// 发送短信
type SendSmsRequest struct {
	Tel  string `json:"tel"`           // 手机号
	Zone string `json:"zone,optional"` // TODO 手机区号
	// TODO 补充验证码相关结果字段
}

func (r *SendSmsRequest) Validate() error {
	if r == nil {
		return global.ErrArgs
	}

	if !telRegx.MatchString(r.Tel) {
		return global.ErrInvalidTel
	}

	return nil
}

type SendSmsResponse struct {
}

// 手机+短信验证码登录
type SmsCheckInRequest struct {
	Tel      string `json:"tel"`
	Zone     string `json:"zone,optional"`
	Code     string `json:"code"` // 短信验证码
	Platform string `json:"platform"`
	// TODO 其它验证字段
}

func (r *SmsCheckInRequest) Validate() error {
	if r == nil {
		return global.ErrArgs
	}

	if !telRegx.MatchString(r.Tel) {
		return global.ErrInvalidTel
	}

	if !codeRegx.MatchString(r.Code) {
		return global.ErrInvalidSmsCode
	}

	if !SupportedPlatform(r.Platform) {
		return global.ErrInvalidPlatform
	}

	r.Platform = TransformPlatform(r.Platform)

	return nil
}

// 密码登录
type PassCheckInRequest struct {
}

// 登录成功返回结果 需要包含登录成功的用户信息
type CheckInResponse struct {
	Uid       uint64   `json:"uid"`
	Nickname  string   `json:"nickname"`
	StyleSign string   `json:"style_sign"`
	Avatar    string   `json:"avatar"`
	Gender    string   `json:"gender"`
	CreateAt  int64    `json:"create_at"`
	Session   *Session `json:"-"`
}

func NewCheckInResponseFromUserInfo(u *UserInfo) *CheckInResponse {
	return &CheckInResponse{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		StyleSign: u.StyleSign,
		Gender:    u.Gender,
	}
}
