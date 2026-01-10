package global

import (
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/note/internal/config"
)

var (
	hostname string
	localIp  string
)

func MustInit(conf *config.Config) {
	hostname = utils.MustGetHostname()
	localIp = utils.MustGetLocalIP()
}

func GetHostname() string {
	return hostname
}

func GetLocalIp() string {
	return localIp
}
