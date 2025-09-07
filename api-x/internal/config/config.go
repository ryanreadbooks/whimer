package config

import (
	"time"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

var (
	Conf Config
)

// 各服务配置
type Config struct {
	Http  rest.RestConf   `json:"http"`
	Redis redis.RedisConf `json:"redis"`

	Backend struct {
		Note     xconf.Discovery `json:"note"`
		Comment  xconf.Discovery `json:"comment"`
		Passport xconf.Discovery `json:"passport"`
		Relation xconf.Discovery `json:"relation"`
		Msger    xconf.Discovery `json:"msger"`
		Search   xconf.Discovery `json:"search"`
	} `json:"backend"`

	Obfuscate struct {
		Note ObfuscateConfig `json:"note"`
		Tag  ObfuscateConfig `json:"tag"`
	} `json:"obfuscate"`

	JobConfig struct {
		NoteEventJob NoteEventJob `json:"note_event_job"`
	} `json:"job_config"`
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

type NoteEventJob struct {
	Interval  time.Duration `json:"interval,default=10s"`
	NumOfList uint32        `json:"num_of_list,default=6"`
}
