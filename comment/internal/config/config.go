package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	External struct {
		Grpc struct {
			Passport string `json:"passport"`
			Seqer    string `json:"seqer"`
			Note string `json:"note"`
		} `json:"grpc"`
	} `json:"external"`
}
