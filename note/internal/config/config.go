package config

import (
	"time"

	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Redis redis.RedisConf `json:"redis"`

	Oss Oss `json:"oss"`

	UploadAuthSign struct {
		JwtId       string        `json:"jwt_id"`
		JwtIssuer   string        `json:"jwt_issuer"`
		JwtSubject  string        `json:"jwt_subject"`
		JwtDuration time.Duration `json:"jwt_duration"`
		Ak          string        `json:"ak"`
		Sk          string        `json:"sk"`
	} `json:"upload_auth_sign"`

	ImgProxyAuth imgproxy.Auth `json:"img_proxy_auth"`

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Counter  xconf.Discovery `json:"counter"`
			Comment  xconf.Discovery `json:"comment"`
			Search   xconf.Discovery `json:"search"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`
}

func (c *Config) Init() error {
	return c.ImgProxyAuth.Init()
}

type Oss struct {
	Endpoint        string `json:"endpoint"`
	Location        string `json:"location"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
	DisplayEndpoint string `json:"display_endpoint"`
	UploadEndpoint  string `json:"upload_endpoint"`
}

func (c *Oss) DisplayEndpointBucket() string {
	return c.DisplayEndpoint + "/" + c.Bucket
}

func (c *Oss) UploadEndpointBucket() string {
	return c.UploadEndpoint + "/" + c.Bucket
}
