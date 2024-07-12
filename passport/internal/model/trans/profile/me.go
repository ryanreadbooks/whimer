package profile

import (
	"unicode/utf8"

	"github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

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
	if nickLen > model.MaxNicknameLen {
		return global.ErrNickNameTooLong
	}

	if nickLen <= 0 {
		return global.ErrNicknameTooShort
	}

	if utf8.RuneCountInString(r.StyleSign) > model.MaxStyleSignLen {
		return global.ErrStyleSignTooLong
	}

	if _, ok := model.GenderMap[r.Gender]; !ok {
		return global.ErrInvalidGender
	}

	return nil
}
