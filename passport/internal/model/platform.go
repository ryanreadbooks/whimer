package model

import "strings"

// 支持的平台
const (
	Web = "web"
)

var (
	platforms = map[string]struct{}{
		Web: {},
	}
)

func SupportedPlatform(p string) bool {
	pp := strings.ToLower(p)
	_, ok := platforms[pp]
	return ok
}

func TransformPlatform(p string) string {
	return strings.ToLower(p)
}
