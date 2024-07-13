package oss

// 通用的对象存储配置
type Conf struct {
	User     string `json:"user"`
	Pass     string `json:"pass"`
	Endpoint string `json:"endpoint"`
	Location string `json:"location"`
	Bucket   string `json:"bucket"`
	Prefix   string `json:"prefix"`
}
