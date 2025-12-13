package config

import (
	"fmt"
	"net/url"
	"time"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"

	"github.com/zeromicro/go-zero/core/discov"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Http rest.RestConf      `json:"http"`
	Grpc zrpc.RpcServerConf `json:"grpc"`
	Log  logx.LogConf       `json:"log"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	External struct {
		Grpc struct {
			Passport  xconf.Discovery `json:"passport"`
			Counter   xconf.Discovery `json:"counter"`
			Comment   xconf.Discovery `json:"comment"`
			Search    xconf.Discovery `json:"search"`
			Conductor xconf.Discovery `json:"conductor"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`

	DevCallbacks DevCallbacks `json:"dev_callbacks"`

	RetryConfig RetryConfig `json:"retry_config"`

	// 额外使用etcd进行分片分配
	Etcd discov.EtcdConf `json:"etcd"`
}

func (c *Config) Init() error {
	_, err := url.Parse(c.DevCallbacks.NoteProcessCallback)
	if err != nil {
		return fmt.Errorf("dev_callbacks.note_process_callback is not a valid url: %w", err)
	}

	return nil
}

type DevCallbacks struct {
	NoteProcessCallback string `json:"note_process_callback"`
}

type RetryConfig struct {
	ProcedureRetry struct {
		TaskRegister struct {
			ScanInterval   time.Duration `json:"scan_interval,default=10s"`
			RetryInterval  time.Duration `json:"retry_interval,default=1m"`
			Limit          int           `json:"limit,default=200"`
			SlotGapSec     int           `json:"slot_gap_sec,default=10"` // 时间片长度 单位秒
			FutureInterval time.Duration `json:"future_interval,default=1m"`
		} `json:"task_register"`
	} `json:"procedure_retry"`
}
