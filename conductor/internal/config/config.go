package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`
	Log  logx.LogConf       `json:"log"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	Kafka *KafkaConfig `json:"kafka"`

	// 额外使用etcd进行分片分配
	Etcd discov.EtcdConf `json:"etcd"`

	// 分片配置
	Shard ShardConfig `json:"shard"`
}

type ShardConfig struct {
	// 抢占重试间隔，默认 500ms
	ClaimRetryInterval time.Duration `json:"claim_retry_interval,omitempty"`
	// 定期检查间隔，默认 5s
	CheckInterval time.Duration `json:"check_interval,omitempty"`
}

func (c *ShardConfig) GetClaimRetryInterval() time.Duration {
	if c.ClaimRetryInterval <= 0 {
		return 500 * time.Millisecond
	}
	return c.ClaimRetryInterval
}

func (c *ShardConfig) GetCheckInterval() time.Duration {
	if c.CheckInterval <= 0 {
		return 5 * time.Second
	}
	return c.CheckInterval
}

type KafkaConfig struct {
	Brokers  string `json:"brokers"`
	Username string `json:"username"`
	Password string `json:"password"`
}
