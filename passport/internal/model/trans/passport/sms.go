package passport

import (
	"regexp"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model/platform"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
)

var (
	telRegx  = regexp.MustCompile(`^1[3-9]\d{9}$`)
	codeRegx = regexp.MustCompile(`^[1-9][0-9]{5}$`)
)

type SmsSendReq struct {
	Tel  string `json:"tel"`           // 手机号
	Zone string `json:"zone,optional"` // TODO 手机区号
	// TODO 补充验证码相关结果字段
}

func (r *SmsSendReq) Validate() error {
	if r == nil {
		return global.ErrArgs
	}

	if !telRegx.MatchString(r.Tel) {
		return global.ErrInvalidTel
	}

	return nil
}

type SmdSendRes struct {
}

// 手机+短信验证码登录
type SignInSmdReq struct {
	Tel      string `json:"tel"`
	Zone     string `json:"zone,optional"`
	Code     string `json:"code"` // 短信验证码
	Platform string `json:"platform"`
	// TODO 其它验证字段
}

func (r *SignInSmdReq) Validate() error {
	if r == nil {
		return global.ErrArgs
	}

	if !telRegx.MatchString(r.Tel) {
		return global.ErrInvalidTel
	}

	if !codeRegx.MatchString(r.Code) {
		return global.ErrInvalidSmsCode
	}

	if !platform.Supported(r.Platform) {
		return global.ErrInvalidPlatform
	}

	r.Platform = platform.Transform(r.Platform)

	return nil
}

// 登录成功返回结果 需要包含登录成功的用户信息
type SignInSmsRes struct {
	Uid       uint64 `json:"uid"`
	Nickname  string `json:"nickname"`
	StyleSign string `json:"style_sign"`
	Avatar    string `json:"avatar"`
	Gender    int8   `json:"gender"`
	CreateAt  int64  `json:"create_at"`
}

func NewFromRepoBasic(u *userbase.Basic) *SignInSmsRes {
	return &SignInSmsRes{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		StyleSign: u.StyleSign,
		Gender:    u.Gender,
		CreateAt:  u.CreateAt,
	}
}
