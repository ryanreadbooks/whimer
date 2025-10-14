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
		WsLink   xconf.Discovery `json:"wslink"`
	} `json:"backend"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`

	JobConfig struct {
		NoteEventJob NoteEventJob `json:"note_event_job"`
	} `json:"job_config"`
}

type NoteEventJob struct {
	Interval  time.Duration `json:"interval,default=10s"`
	NumOfList uint32        `json:"num_of_list,default=6"`
}
