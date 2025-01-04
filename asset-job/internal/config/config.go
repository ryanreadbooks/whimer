package config

import "github.com/ryanreadbooks/whimer/misc/xconf"

// 全局变量
var Conf Config

type Config struct {
	NoteAssetEventKafka xconf.KfkConf `json:"note_asset_event_kafka"`

	NoteOss struct {
		Ak        string `json:"ak"`
		Sk        string `json:"sk"`
		Endpoint  string `json:"endpoint"`
		Location  string `json:"location"`
		Bucket    string `json:"bucket"`
		PrvBucket string `json:"prv_bucket"`
	} `json:"note_oss"`
}
