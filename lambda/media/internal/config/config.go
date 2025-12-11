package config

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/core/logx"
)

var Conf Config

type Config struct {
	Log logx.LogConf `json:"log"`

	Conductor xconf.Discovery `json:"conductor"`

	Worker struct {
		Concurrency int `json:"concurrency"`
	} `json:"worker"`
}
