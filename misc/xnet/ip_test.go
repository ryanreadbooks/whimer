package xnet

import (
	"testing"
)

func TestIp4Int64(t *testing.T) {
	cc := []string{
		"127.0.0.1",
		"128.23.12.43",
		"12.a.we,",
	}

	for _, c := range cc {
		t.Log(IpAsInt(c))
	}

	ci := []uint32{
		2130706433,
		2148994091,
		12,
		0,
	}

	for _, i := range ci {
		t.Log(IntAsIp(i))
	}
}
