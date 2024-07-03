package config

import (
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	Oss struct {
		User     string `json:"user"`
		Pass     string `json:"pass"`
		Endpoint string `json:"endpoint"`
		Location string `json:"location"`
		Bucket   string `json:"bucket"`
		Prefix   string `json:"prefix"`
	}

	Idgen struct {
		Addr string `json:"addr"`
	} `json:"idgen"`
}
