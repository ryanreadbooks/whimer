package config

import (
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局配置对象
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

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Note     xconf.Discovery `json:"note"`
			Counter  xconf.Discovery `json:"counter"`
		} `json:"grpc"`
	} `json:"external"`

	Seqer Seqer           `json:"seqer"`
	Redis redis.RedisConf `json:"redis"`

	Cron struct {
	} `json:"cron"`

	Oss struct {
		InlineImage OssConfig `json:"inline_image"`
	} `json:"oss"`

	OssUploadAuth signer.JwtSignConfig `json:"oss_upload_auth"`
	ImgProxyAuth  imgproxy.Auth        `json:"img_proxy_auth"`
}

func (c *Config) Init() error {
	return c.ImgProxyAuth.Init()
}

type Seqer struct {
	Addr string `json:"addr"`
}

type OssConfig struct {
	Endpoint        string `json:"endpoint"`
	Location        string `json:"location"`
	Bucket          string `json:"bucket"`
	Prefix          string `json:"prefix"`
	DisplayEndpoint string `json:"display_endpoint"`
	UploadEndpoint  string `json:"upload_endpoint"`
}

func (c *OssConfig) DisplayEndpointBucket() string {
	return c.DisplayEndpoint + "/" + c.Bucket
}

func (c *OssConfig) UploadEndpointBucket() string {
	return c.UploadEndpoint + "/" + c.Bucket
}
