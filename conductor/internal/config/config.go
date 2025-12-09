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

	// 扫描配置
	ScanConfig ScanConfig `json:"scan_config"`

	// Worker 配置
	WorkerConfig WorkerConfig `json:"worker_config"`
}

type ScanConfig struct {
	// 任务分发扫描间隔，默认 1s
	ProcessInterval time.Duration `json:"process_interval,omitempty"`
	// 重试扫描间隔，默认 5s
	RetryInterval time.Duration `json:"retry_interval,omitempty"`
	// 过期扫描间隔，默认 10s
	ExpireInterval time.Duration `json:"expire_interval,omitempty"`
}

func (c *ScanConfig) GetProcessInterval() time.Duration {
	if c.ProcessInterval <= 0 {
		return 1 * time.Second
	}
	return c.ProcessInterval
}

func (c *ScanConfig) GetRetryInterval() time.Duration {
	if c.RetryInterval <= 0 {
		return 5 * time.Second
	}
	return c.RetryInterval
}

func (c *ScanConfig) GetExpireInterval() time.Duration {
	if c.ExpireInterval <= 0 {
		return 10 * time.Second
	}
	return c.ExpireInterval
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

type WorkerConfig struct {
	// 长轮询超时时间，默认 30s
	LongPollTimeout time.Duration `json:"long_poll_timeout,omitempty"`
}

func (c *WorkerConfig) GetLongPollTimeout() time.Duration {
	if c.LongPollTimeout <= 0 {
		return 30 * time.Second
	}
	return c.LongPollTimeout
}

type KafkaConfig struct {
	Brokers  string `json:"brokers"`
	Username string `json:"username"`
	Password string `json:"password"`
}
