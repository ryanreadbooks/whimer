package imgproxy

import (
	"encoding/hex"
	"testing"
)

func TestDecodeString(t *testing.T) {
	kb, e := hex.DecodeString("9083918dccbb")
	t.Log(kb)
	t.Log(e)
}

func TestGetPublicUrl(t *testing.T) {
	res := GetSignedUrlWith("127.0.0.1:10000", "nota/rrv10",
		"dbf5f7eedf9a04309a324ac63c6d01294d18e392",
		"29868e116f43a735ad58cf963a8e9dd81779a796",
		WithQuality("36"))
	t.Log(res)

	res = GetSignedUrlWith("127.0.0.1:10000", "nota/rrv10",
		"dbf5f7eedf9a04309a324ac63c6d01294d18e392",
		"29868e116f43a735ad58cf963a8e9dd81779a796")
	t.Log(res)

	res = GetSignedUrlWith("127.0.0.1:10000", "pics/cmt_inline/abcedfe",
		"dbf5f7eedf9a04309a324ac63c6d01294d18e392",
		"29868e116f43a735ad58cf963a8e9dd81779a796")
	t.Log(res)
}
