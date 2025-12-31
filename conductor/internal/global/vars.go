package global

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/misc/utils"
)

var (
	hostname string
	localIp  string

	ipAndPort string
)

func MustInit(conf *config.Config) {
	hostname = utils.MustGetHostname()
	localIp = utils.MustGetLocalIP()

	ss := strings.SplitN(conf.Grpc.ListenOn, ":", 2)
	grpcPort, _ := strconv.Atoi(ss[1])
	ipAndPort = fmt.Sprintf("%s:%d", localIp, grpcPort)
}

func GetIpAndPort() string {
	return ipAndPort
}

func GetHostname() string {
	return hostname
}

func GetLocalIp() string {
	return localIp
}
