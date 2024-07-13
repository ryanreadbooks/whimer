package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Http rest.RestConf      `json:"http"`
	Grpc zrpc.RpcServerConf `json:"grpc"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	Oss struct {
		User     string `json:"user"`
		Pass     string `json:"pass"`
		Endpoint string `json:"endpoint"`
		Location string `json:"location"`
		Bucket   string `json:"bucket"`
		Prefix   string `json:"prefix"`
	}

	Idgen struct {
		Addr string `json:"addr"`
	} `json:"idgen"`
}
