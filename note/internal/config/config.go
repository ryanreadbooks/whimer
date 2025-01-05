package config

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

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
		User            string `json:"user"`
		Pass            string `json:"pass"`
		Endpoint        string `json:"endpoint"`
		Location        string `json:"location"`
		Bucket          string `json:"bucket"`
		BucketPreview   string `json:"bucket_prv"`
		Prefix          string `json:"prefix"`
		DisplayEndpoint string `json:"display_endpoint"`
	} `json:"oss"`

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Counter  xconf.Discovery `json:"counter"`
			Comment  xconf.Discovery `json:"comment"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`
}
