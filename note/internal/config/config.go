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

	Oss struct {
		User            string `json:"user"`
		Pass            string `json:"pass"`
		Endpoint        string `json:"endpoint"`
		Location        string `json:"location"`
		Bucket          string `json:"bucket"`
		Prefix          string `json:"prefix"`
		DisplayEndpoint string `json:"display_endpoint"`
	} `json:"oss"`

	External struct {
		Grpc struct {
			Passport string `json:"passport"`
		} `json:"grpc"`
	} `json:"3rd"`
}
