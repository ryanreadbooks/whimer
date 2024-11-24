package config

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/rest"
)

var Conf Config

type Config struct {
	Http rest.RestConf `json:"http"`

	Backend struct {
		Note     xconf.Discovery `json:"note"`
		Comment  xconf.Discovery `json:"comment"`
		Passport xconf.Discovery `json:"passport"`
	} `json:"backend"`
}
