package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	ElasticSearch ElasticSearch `json:"elasticsearch"`
}

type ElasticSearch struct {
	Addr     string `json:"addr"` // comma seperated addresses
	User     string `json:"user"`
	Password string `json:"password"`
}
