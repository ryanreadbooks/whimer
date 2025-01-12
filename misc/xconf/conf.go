package xconf

import (
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/zrpc"
)

type Discovery struct {
	Hosts []string `json:"hosts"`
	Key   string   `json:"key"`
}

func (c Discovery) AsZrpcClientConf() zrpc.RpcClientConf {
	return zrpc.RpcClientConf{
		Etcd: discov.EtcdConf{
			Hosts: c.Hosts,
			Key:   c.Key,
		},
	}
}

type KfkConf struct {
	Brokers      []string `json:"brokers"`
	Topic        string   `json:"topic"`
	ConsumeGroup string   `json:"consume_group"`
	NumConsumers int      `json:"num_consumers"`
	Offset       string   `json:"offset"`
}

func (kc KfkConf) AsKqConf() kq.KqConf {
	return kq.KqConf{
		Brokers:       kc.Brokers,
		Topic:         kc.Topic,
		Group:         kc.ConsumeGroup,
		Consumers:     kc.NumConsumers,
		Offset:        kc.Offset,
		CommitInOrder: true,
	}
}

type KfkNotifyConf struct {
	Brokers []string `json:"brokers"`
	Topic   string   `json:"topic"`
}
