package dep

import (
	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

// 调度服务
var (
	conductProducer *producer.Client
)

func InitConductor(c *config.Config) {
	cli, _ := producer.New(producer.ClientOptions{
		HostConf:  c.Backend.Conductor,
		Namespace: "pilot",
	})

	conductProducer = cli
}

func ConductorProducer() *producer.Client {
	return conductProducer
}
