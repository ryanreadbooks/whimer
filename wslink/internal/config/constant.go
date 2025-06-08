package config

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/utils"
)

var (
	Hostname string
	LocalIP  string

	ipAndPort string
)

func Init() {
	Hostname = utils.MustGetHostname()
	LocalIP = utils.MustGetLocalIP()

	ss := strings.SplitN(Conf.Grpc.ListenOn, ":", 2)
	grpcPort, _ := strconv.Atoi(ss[1])
	ipAndPort = fmt.Sprintf("%s:%d", LocalIP, grpcPort)
}

func GetIpAndPort() string {
	return ipAndPort
}
