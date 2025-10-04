package xnet

import (
	"math/big"
	"net"
)

// ipv4转成int
func IpAsInt(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

// int转成ipv4
func IntAsIp(sum uint32) string {
	ip := make(net.IP, net.IPv4len)
	ip[0] = byte((sum >> 24) & 0xFF)
	ip[1] = byte((sum >> 16) & 0xFF)
	ip[2] = byte((sum >> 8) & 0xFF)
	ip[3] = byte(sum & 0xFF)
	return ip.String()
}

func IpAsBytes(ip string) []byte {
	return net.ParseIP(ip)
}

func BytesIpAsString(ip []byte) string {
	if len(ip) == 0 {
		return ""
	}
	return net.IP(ip).String()
}
