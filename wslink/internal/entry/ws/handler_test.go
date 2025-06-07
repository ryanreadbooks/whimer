package ws

import (
	"encoding/base64"
	"encoding/hex"
	"testing"

	v1 "github.com/ryanreadbooks/whimer/wslink/api/protocol/v1"
	protobuf "google.golang.org/protobuf/proto"
)

func Test_GenPingWire(t *testing.T) {
	wire := v1.Protocol{
		Meta: &v1.Meta{
			Flag: v1.Flag_FLAG_PING,
		},
	}

	// CgIIAg==
	data, _ := protobuf.Marshal(&wire)
	dd := base64.StdEncoding.EncodeToString(data)
	t.Log(dd)
	t.Log(hex.EncodeToString(data))
}

func Test_GenDataWire(t *testing.T) {
	wire := v1.Protocol{
		Meta: &v1.Meta{
			Flag: v1.Flag_FLAG_DATA,
		},
		Payload: []byte("HELLO WSLINK"),
	}

	// CgIIAxIMSEVMTE8gV1NMSU5L
	data, _ := protobuf.Marshal(&wire)
	dd := base64.StdEncoding.EncodeToString(data)
	t.Log(dd)
}

func Test_GenData(t *testing.T) {
	var s = "string is the way to do it"
	b := []byte(s)
	t.Log(base64.StdEncoding.EncodeToString(b))
}
