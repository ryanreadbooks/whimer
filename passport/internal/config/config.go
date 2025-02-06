package config

import (
	"github.com/ryanreadbooks/whimer/misc/imgproxy"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Http   rest.RestConf      `json:"http"`
	Grpc   zrpc.RpcServerConf `json:"grpc"`
	Domain string             `json:"domain"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	Oss Oss `json:"oss"`

	ImgProxyAuth imgproxy.Auth `json:"img_proxy_auth"`

	Idgen struct {
		Addr string `json:"addr"`
	} `json:"idgen"`
}

func (c *Config) Init() error {
	return c.ImgProxyAuth.Init()
}

type Oss struct {
	Ak              string `json:"ak"`
	Sk              string `json:"sk"`
	Endpoint        string `json:"endpoint"`
	Location        string `json:"location"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
	DisplayEndpoint string `json:"display_endpoint"`
}

func (c *Oss) AvatarDisplayEndpoint() string {
	return c.DisplayEndpoint + "/" + "avatar"
}
