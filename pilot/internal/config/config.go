package config

import (
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

var (
	Conf Config
)

// 各服务配置
type Config struct {
	Http  rest.RestConf   `json:"http"`
	Redis redis.RedisConf `json:"redis"`
	Log   logx.LogConf    `json:"log"`

	Backend struct {
		Note     xconf.Discovery `json:"note"`
		Comment  xconf.Discovery `json:"comment"`
		Passport xconf.Discovery `json:"passport"`
		Relation xconf.Discovery `json:"relation"`
		Msger    xconf.Discovery `json:"msger"`
		Search   xconf.Discovery `json:"search"`
		WsLink   xconf.Discovery `json:"wslink"`
	} `json:"backend"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`

	JobConfig struct {
		NoteEventJob NoteEventJob `json:"note_event_job"`
	} `json:"job_config"`

	Kafka *KafkaConfig `json:"kafka"`

	UploadAuthSign UploadAuthSign `json:"upload_auth_sign"`

	UploadResourceDefineMap UploadResourceDefineMap     `json:"-"`
	UploadResourceDefine    []*UploadResourceDefineItem `json:"upload_resource_define"`

	Oss Oss `json:"oss"`

	ImgProxyAuth imgproxy.Auth `json:"img_proxy_auth"`
}

func (c *Config) Init() error {
	err := c.ImgProxyAuth.Init()
	if err != nil {
		return err
	}

	// init upload define map
	c.UploadResourceDefineMap = make(UploadResourceDefineMap)
	for _, item := range c.UploadResourceDefine {
		if !uploadresource.CheckValid(item.Name) {
			return fmt.Errorf("uploadresource %s invalid", item.Name)
		}
		c.UploadResourceDefineMap[uploadresource.Type(item.Name)] = item.Meta
	}

	return nil
}

type NoteEventJob struct {
	Interval  time.Duration `json:"interval,default=10s"`
	NumOfList uint32        `json:"num_of_list,default=6"`
}

type KafkaConfig struct {
	Brokers  string `json:"brokers"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UploadAuthSign struct {
	JwtId       string        `json:"jwt_id"`
	JwtIssuer   string        `json:"jwt_issuer"`
	JwtSubject  string        `json:"jwt_subject"`
	JwtDuration time.Duration `json:"jwt_duration"`
	Ak          string        `json:"ak"`
	Sk          string        `json:"sk"`
}

type UploadResourceDefineItem struct {
	Name string                  `json:"name"`
	Meta uploadresource.Metadata `json:"meta"`
}

type UploadResourceDefineMap map[uploadresource.Type]uploadresource.Metadata // string is resourceType

type Oss struct {
	Endpoint        string `json:"endpoint"`
	Location        string `json:"location"`
	DisplayEndpoint string `json:"display_endpoint"`
	UploadEndpoint  string `json:"upload_endpoint"`
}

func (c *Oss) DisplayEndpointBucket(bucket string) string {
	return c.DisplayEndpoint + "/" + bucket
}

func (c *Oss) UploadEndpointBucket(bucket string) string {
	return c.UploadEndpoint + "/" + bucket
}
