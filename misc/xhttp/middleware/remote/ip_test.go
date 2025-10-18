package remote

import (
	"net"
	"testing"
)

func TestTcpAddrParse(t *testing.T) {
	host, port,err := net.SplitHostPort("192.168.3.1:9000")
	t.Log(err)
	t.Log(host, port)

	host, port , err = net.SplitHostPort("[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:9000")
	t.Log(err)
	t.Log(host, port)

	t.Log(net.ParseIP(""))
	t.Log(net.IP([]byte{}).String())
}