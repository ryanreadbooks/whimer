package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	Http rest.RestConf `json:"http"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	ThreeRd struct {
		Grpc struct {
			Passport string `json:"passport"`
		} `json:"grpc"`
	} `json:"3rd"`
}
