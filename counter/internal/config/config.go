package config

import (
	"os"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	Cron struct {
		SyncerSpec string `json:"syncer_spec"`
	} `json:"cron"`
}

func ConfigForTest() *Config {
	c := &Config{}
	c.MySql.User = os.Getenv("ENV_DB_USER")
	c.MySql.Pass = os.Getenv("ENV_DB_PASS")
	c.MySql.Addr = os.Getenv("ENV_DB_ADDR")
	c.MySql.DbName = os.Getenv("ENV_DB_NAME")

	return c
}
