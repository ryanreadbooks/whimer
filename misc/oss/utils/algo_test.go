package utils

import (
	"encoding/hex"
	"net/http"
	"testing"
)

func TestAlgo(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/nota/rrv10", nil)
	req.Header.Set("Host", "s1-file.whimer.com")
	req.Header.Set("X-Date", "20250203T085400Z")
	req.Header.Set("Content-Length", "2035229")
	req.Header.Set("X-Security-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ3aG1fbm90ZSIsInN1YiI6InN0cyIsImV4cCI6MTczODU5Mjk0NSwibmJmIjoxNzM4NTg5MzQ1LCJpYXQiOjE3Mzg1ODkzNDUsImp0aSI6IndobV91bGFzIiwiYWNjZXNzX2tleSI6ImMyYzVmM2Y4ZjVlMjAyMTQ5YzI5MWNhMWY4NDRkZjQzNjU3ZTQ5NjAiLCJyZXNvdXJjZSI6ImltYWdlIiwic291cmNlIjoid2ViIn0.5IFdveMYPjepTdyH__xRavKX4UUbNezjMoBsq-8D2Mk")

	res, _ := WhmrSign("c2c5f3f8f5e202149c291ca1f844df43657e4960", req)
	t.Log(res)
}

func TestSum(t *testing.T) {
	b := sumHMAC([]byte("79ecc1e862c715c1676659093f479608bbb24b28"), []byte("abc"))
	t.Log(b)
	t.Log(hex.EncodeToString(b))
}
