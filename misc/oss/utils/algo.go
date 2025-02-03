package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7/pkg/s3utils"
)

func sumHMAC(key, data []byte) []byte {
	hash := hmac.New(sha256.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}

// 简单签名一个请求
func WhmrSign(secret string, req *http.Request) (string, error) {
	ctl := req.Header.Get("Content-Length")
	host := req.Header.Get("Host")
	xDate := req.Header.Get("X-Date")
	xSecuToken := req.Header.Get("X-Security-Token")
	method := req.Method
	uri := s3utils.EncodePath(req.URL.Path)

	var builder strings.Builder
	builder.WriteString(method)
	builder.WriteByte('\n')
	builder.WriteString(uri)
	builder.WriteByte('\n')
	builder.WriteString("content-length:" + ctl + "\n")
	builder.WriteString("host:" + host + "\n")
	builder.WriteString("x-date:" + xDate + "\n")
	builder.WriteString("x-security-token:" + xSecuToken + "\n")
	builder.WriteByte('\n')
	builder.WriteString("content-length;host;x-date;x-security-token")
	builder.WriteByte('\n')

	hashedCanonicalReq := hex.EncodeToString(sumHMAC([]byte(secret), []byte(builder.String())))

	stringToSign := "SHA256" + "\n" + xDate + "\n" + hashedCanonicalReq
	seck := "WHMR" + secret
	signingKey := sumHMAC([]byte(seck), []byte(xDate))
	signingKey = sumHMAC(signingKey, []byte("whmr_request"))
	println(hex.EncodeToString(signingKey))
	signature := sumHMAC(signingKey, []byte(stringToSign))

	return hex.EncodeToString(signature), nil
}
