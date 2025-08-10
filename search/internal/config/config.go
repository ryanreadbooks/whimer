package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	ElasticSearch ElasticSearch `json:"elasticsearch"`
	Indices       struct {
		NoteTag Index `json:"note_tag"`
	} `json:"indices"`
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
