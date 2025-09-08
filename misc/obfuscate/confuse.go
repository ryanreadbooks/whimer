package obfuscate

import (
	"fmt"
	"strconv"

	"github.com/speps/go-hashids/v2"
)

const (
	salt    = "this is whimer misc default salt"
	iMinLen = 20
)

var (
	ErrDemixEmpty = fmt.Errorf("demix result empty")
)

var (
	// default settings
	hd = &hashids.HashIDData{
		Alphabet:  hashids.DefaultAlphabet,
		Salt:      salt,
		MinLength: iMinLen,
	}
)

func Mix(number int64) (string, error) {
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return "", err
	}
	return h.EncodeInt64([]int64{number})
}

func DeMix(s string) (int64, error) {
	h, err := hashids.NewWithData(hd)
	if err != nil {
		return 0, err
	}

	res, err := h.DecodeInt64WithError(s)
	if err != nil {
		return 0, nil
	}
	if len(res) <= 0 {
		return 0, ErrDemixEmpty
	}

	return res[0], nil
}

type Obfuscate interface {
	Mix(int64) (string, error)
	DeMix(string) (int64, error)

	MixU(uint64) (string, error)
	DeMixU(string) (uint64, error)
}

type Confuser struct {
	hd *hashids.HashIDData
	h  *hashids.HashID
}

type Option func(*hashids.HashIDData)

func WithSalt(salt string) Option {
	return func(hi *hashids.HashIDData) {
		hi.Salt = salt
	}
}

func WithMinLen(l int) Option {
	return func(hi *hashids.HashIDData) {
		hi.MinLength = l
	}
}

func WithAlphabet(s string) Option {
	return func(hi *hashids.HashIDData) {
		hi.Alphabet = s
	}
}

type Config struct {
	Salt      string `json:"salt"`
	MinLength int    `json:"min_length,default=12"`
	Alphabet  string `json:"alphabet,optional"`
}

func (c *Config) Options() []Option {
	opts := []Option{
		WithSalt(c.Salt),
		WithMinLen(c.MinLength),
	}

	if c.Alphabet != "" {
		opts = append(opts, WithAlphabet(c.Alphabet))
	}

	return opts
}

func NewConfuser(opts ...Option) (*Confuser, error) {
	hd := &hashids.HashIDData{
		Alphabet:  hashids.DefaultAlphabet,
		Salt:      salt,
		MinLength: iMinLen,
	}

	for _, o := range opts {
		o(hd)
	}

	h, err := hashids.NewWithData(hd)
	if err != nil {
		return nil, err
	}

	return &Confuser{
		hd: hd,
		h:  h,
	}, nil
}

func (c *Confuser) Mix(number int64) (string, error) {
	return c.h.EncodeInt64([]int64{number})
}

func (c *Confuser) DeMix(s string) (int64, error) {
	res, err := c.h.DecodeInt64WithError(s)
	if err != nil {
		return 0, err
	}
	if len(res) <= 0 {
		return 0, ErrDemixEmpty
	}

	return res[0], nil
}

func (c *Confuser) MixU(number uint64) (string, error) {
	s, _ := c.h.EncodeHex(strconv.FormatUint(number, 10))
	return s, nil
}

func (c *Confuser) DeMixU(s string) (uint64, error) {
	res, err := c.h.DecodeHex(s)
	if err != nil {
		return 0, err
	}
	if len(res) <= 0 {
		return 0, ErrDemixEmpty
	}
	number, err := strconv.ParseUint(res, 10, 64)
	if err != nil {
		return 0, err
	}

	return number, nil
}
