package infra

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xnet"
)

var (
	ipRegionConverter xnet.IpRegionConverter
)

func initMisc(c *config.Config) {
	ipRegionConverter = xnet.NewUnknownIpRegionConverter()
}

func Ip2Loc() xnet.IpRegionConverter {
	return ipRegionConverter
}
