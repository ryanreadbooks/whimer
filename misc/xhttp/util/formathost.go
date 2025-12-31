package util

import "strings"

func FormatHost(host string, secure bool) string {
	if secure {
		if !strings.HasPrefix(host, "https://") {
			return "https://" + host
		}
	} else {
		if !strings.HasPrefix(host, "http://") {
			return "http://" + host
		}
	}
	return host
}
