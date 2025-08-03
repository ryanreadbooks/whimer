package config

import (
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/rest"
)

// 各服务配置
type Config struct {
	Http rest.RestConf `json:"http"`

	Backend struct {
		Note     xconf.Discovery `json:"note"`
		Comment  xconf.Discovery `json:"comment"`
		Passport xconf.Discovery `json:"passport"`
		Relation xconf.Discovery `json:"relation"`
		Msger    xconf.Discovery `json:"msger"`
	} `json:"backend"`

	Obfuscate struct {
		Note ObfuscateConfig `json:"note"`
		Tag  ObfuscateConfig `json:"tag"`
	} `json:"obfuscate"`
}

type ObfuscateConfig struct {
	Salt      string `json:"salt"`
	MinLength int    `json:"min_length,default=12"`
	Alphabet  string `json:"alphabet,optional"`
}

func (c *ObfuscateConfig) Options() []obfuscate.Option {
	opts := []obfuscate.Option{
		obfuscate.WithSalt(c.Salt),
		obfuscate.WithMinLen(c.MinLength),
	}

	if c.Alphabet != "" {
		opts = append(opts, obfuscate.WithAlphabet(c.Alphabet))
	}

	return opts
}
