package safety

import (
	"github.com/speps/go-hashids/v2"
)

const (
	salt = "this is whimer llddzzrrbaba"

	iMinLen = 24
)

var (
	hd = &hashids.HashIDData{
		Alphabet:  hashids.DefaultAlphabet,
		Salt:      salt,
		MinLength: iMinLen,
	}
)

func Confuse(number int64) string {
	h, _ := hashids.NewWithData(hd)
	s, _ := h.EncodeInt64([]int64{number})

	return s
}

func DeConfuse(s string) int64 {
	h, _ := hashids.NewWithData(hd)
	res, err := h.DecodeInt64WithError(s)
	if err != nil || len(res) <= 0 {
		return 0
	}

	return res[0]
}

type Confuser struct {
	hd *hashids.HashIDData
}

func NewConfuser(salt string, minLen int) *Confuser {
	return &Confuser{
		hd: &hashids.HashIDData{
			Alphabet:  hashids.DefaultAlphabet,
			Salt:      salt,
			MinLength: minLen,
		},
	}
}

func (c *Confuser) Confuse(number int64) string {
	h, _ := hashids.NewWithData(c.hd)
	s, _ := h.EncodeInt64([]int64{number})
	return s
}

func (c *Confuser) DeConfuse(s string) int64 {
	h, _ := hashids.NewWithData(c.hd)
	res, err := h.DecodeInt64WithError(s)
	if err != nil || len(res) <= 0 {
		return 0
	}

	return res[0]
}
