package util

import (
	"strings"

	"google.golang.org/grpc"
)

func SplitUnaryServerName(info *grpc.UnaryServerInfo) string {
	seps := strings.SplitN(info.FullMethod, "/", 3)
	if len(seps) != 3 {
		return ""
	}
	return seps[1]
}
