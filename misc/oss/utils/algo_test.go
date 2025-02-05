package utils

import (
	"encoding/hex"
	"net/http"
	"testing"
)

func TestAlgo(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/nota/c3d20edd9c6c56377524dbbf359c91fb70396734", nil)
	req.Header.Set("Host", "s1-upload.whimer.com")
	req.Header.Set("X-Date", "20250205T145826Z")
	req.Header.Set("Content-Length", "1467709")
	req.Header.Set("X-Security-Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ3aG1fbm90ZSIsInN1YiI6InN0cyIsImV4cCI6MTczODc3NDcwMywibmJmIjoxNzM4NzY3NTAzLCJpYXQiOjE3Mzg3Njc1MDMsImp0aSI6IndobV91bGFzIiwiYWNjZXNzX2tleSI6IjQyMzIwNzI5MjkwMDNmMjI3NjhlNTA5OWNmYjlmNjNmZThmN2M4NmMiLCJyZXNvdXJjZSI6ImltYWdlIiwic291cmNlIjoid2ViIn0.ru3PGBJQAWlt-JP-jwDSWJmjTSvpK9KR2N_rcJhgYCw")

	res, _ := WhmrSign("4232072929003f22768e5099cfb9f63fe8f7c86c", req)
	t.Log(res)
}

func TestSum(t *testing.T) {
	b := sumHMAC([]byte("391e46b7cfd315c2c0a86a62c67330a7fb829d0e"), []byte("abc"))
	t.Log(b)
	t.Log(hex.EncodeToString(b))
}
