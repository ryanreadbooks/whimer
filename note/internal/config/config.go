package config

import (
	"encoding/hex"
	"fmt"
	"time"

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

	Oss Oss `json:"oss"`

	UploadAuthSign struct {
		JwtId       string        `json:"jwt_id"`
		JwtIssuer   string        `json:"jwt_issuer"`
		JwtSubject  string        `json:"jwt_subject"`
		JwtDuration time.Duration `json:"jwt_duration"`
		Ak          string        `json:"ak"`
		Sk          string        `json:"sk"`
	} `json:"upload_auth_sign"`

	ImgProxyAuth ImgProxyAuth `json:"img_proxy_auth"`

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Counter  xconf.Discovery `json:"counter"`
			Comment  xconf.Discovery `json:"comment"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`
}

func (c *Config) Init() error {
	return c.ImgProxyAuth.Init()
}

type ImgProxyAuth struct {
	Key  string `json:"key"`
	Salt string `json:"salt"`

	keyBin  []byte `json:"-" yaml:"-"`
	saltBin []byte `json:"-" yaml:"-"`
}

func (c *ImgProxyAuth) GetKey() []byte {
	return c.keyBin
}

func (c *ImgProxyAuth) GetSalt() []byte {
	return c.saltBin
}

func (c *ImgProxyAuth) Init() error {
	var err error
	c.keyBin, err = hex.DecodeString(c.Key)
	if err != nil {
		return fmt.Errorf("img proxy auth key is invalid: %w", err)
	}

	c.saltBin, err = hex.DecodeString(c.Salt)
	if err != nil {
		return fmt.Errorf("img proxy auth salt is invalid: %w", err)
	}

	return nil
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
