package xnet

import "context"

type IpRegionConverter interface {
	Convert(ctx context.Context, p string) (string, error)
}

type unknownRegionConverter struct{}

func (unknownRegionConverter) Convert(ctx context.Context, ip string) (string, error) {
	return "未知", nil
}

func NewUnknownIpRegionConverter() IpRegionConverter {
	return unknownRegionConverter{}
}
