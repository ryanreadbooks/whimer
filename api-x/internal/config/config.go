package config

import (
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
	} `json:"obfuscate"`
}

type ObfuscateConfig struct {
	Salt      string `json:"salt"`
	MinLength int    `json:"min_length,default=12"`
}
