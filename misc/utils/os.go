package utils

import (
	"fmt"
	"net"
	"os"
)

func GetLocalIP() (string, error) {
	// 拿到本机所有网口
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// 获取接口所有地址
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}

			ip4 := ipNet.IP.To4()
			if ip4 != nil && !ip4.IsLoopback() {
				// 不要多播地址
				if !ip4.IsMulticast() {
					return ip4.String(), nil
				}
			}
		}
	}

	return "", fmt.Errorf("no local IP address found")
}

func MustGetHostname() string {
	h, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	return h
}

func MustGetLocalIP() string {
	ip, err := GetLocalIP()
	if err != nil {
		panic(err)
	}

	return ip
}