package utils

import (
	"encoding/hex"
	"net/http"
	"testing"
)

func TestAlgo(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/nota/rrv9", nil)
	req.Header.Set("Host", "s1-file.whimer.com")
	req.Header.Set("X-Date", "20250203T085400Z")
	req.Header.Set("Content-Length", "2035229")
	req.Header.Set("X-Security-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ3aG1fbm90ZSIsInN1YiI6InN0cyIsImV4cCI6MTczODU3NTkzNiwibmJmIjoxNzM4NTcyMzM2LCJpYXQiOjE3Mzg1NzIzMzYsImp0aSI6IndobV91bGFzIiwiYWNjZXNzX2tleSI6IjI1MDQzYTNjMzlkNTA5OGEyZWU3NjVhZmRlMTMzYWM0MDc5MjU0MWYiLCJyZXNvdXJjZSI6ImltYWdlIiwic291cmNlIjoid2ViIn0.j_vAWQS4kaNrFHPiT8ZRmIi0UNacLU2otjwzN8l6Bkg")

	res, _ := WhmrSign("25043a3c39d5098a2ee765afde133ac40792541f", req)
	t.Log(res)
}

func TestSum(t *testing.T) {
	b := sumHMAC([]byte("79ecc1e862c715c1676659093f479608bbb24b28"), []byte("abc"))
	t.Log(b)
	t.Log(hex.EncodeToString(b))
}
