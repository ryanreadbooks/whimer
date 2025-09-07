package config

import (
	"time"

	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	ElasticSearch ElasticSearch `json:"elasticsearch"`
	Indices       struct {
		NoteTag Index `json:"note_tag"`
		Note    Index `json:"note"`
	} `json:"indices"`

	Kafka struct {
		Brokers  string `json:"brokers"` // comma seperated addresses
		Username string `json:"username"`
		Password string `json:"password"`
	} `json:"kafka"`

	ConsumerTopicConfig ConsumerTopicConfig `json:"consumer_topic_config"`
}

type ConsumerTopicConfig map[string]ConsumerHandleConfig

func (cm ConsumerTopicConfig) Get(topic string) ConsumerHandleConfig {
	v, ok := cm[topic]
	if !ok {
		return ConsumerHandleConfig{BatchSize: 100, BatchTimeout: time.Second * 2}
	}

	return v
}

type ElasticSearch struct {
	Addr     string `json:"addr"` // comma seperated addresses
	User     string `json:"user"`
	Password string `json:"password"`
}

type Index struct {
	NumReplicas int `json:"num_replicas,default=0"`
	NumShards   int `json:"num_shards,default=1"`
}

type ConsumerHandleConfig struct {
	BatchSize    int           `json:"batch_size,default=100"`
	BatchTimeout time.Duration `json:"batch_timeout,default=2s"`
}
