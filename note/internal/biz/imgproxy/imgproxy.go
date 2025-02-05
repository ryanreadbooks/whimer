package imgproxy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"strings"

	"github.com/ryanreadbooks/whimer/note/internal/config"
)

type processOption struct {
	baseUrl string
	quality string
	ext     string
}

func (p *processOption) perform() string {
	var bd strings.Builder
	if p.quality != "" {
		bd.WriteString(p.quality)
	}
	bd.WriteByte('/')
	bd.WriteString(p.baseUrl)
	bd.WriteString(p.ext)

	return bd.String()
}

type ProcessOpt func(*processOption)

func WithWebpExt() ProcessOpt {
	return func(po *processOption) {
		po.ext = ".webp"
	}
}

func WithJpgExt() ProcessOpt {
	return func(po *processOption) {
		po.ext = ".jpg"
	}
}

func WithQuality(q string) ProcessOpt {
	return func(po *processOption) {
		po.quality = "/" + "q:" + q
	}
}

func getUrlToSign(assetKey string, opts ...ProcessOpt) string {
	assetKey = strings.TrimLeft(assetKey, "/")
	canonicalUrl := getCanonicalUrl(assetKey)

	opt := &processOption{baseUrl: canonicalUrl, ext: ".webp"}
	for _, o := range opts {
		o(opt)
	}

	urlToSign := opt.perform()
	if !strings.HasPrefix(urlToSign, "/") {
		urlToSign = "/" + urlToSign
	}

	return urlToSign
}

func pathJoin(host, signature, urlToSign string) string {
	s, _ := url.JoinPath(signature, urlToSign)
	if strings.HasPrefix(s, "/") {
		return host + s
	}
	return host + "/" + s
}

// Generate public url for image asset key
//
// assetKey is like /bucket/keyName
func GetSignedUrl(host, assetKey string, opts ...ProcessOpt) string {
	urlToSign := getUrlToSign(assetKey, opts...)
	signature := signUrl(urlToSign)
	return pathJoin(host, signature, urlToSign)
}

func GetSignedUrlWith(host, assetKey, key, salt string, opts ...ProcessOpt) string {
	urlToSign := getUrlToSign(assetKey, opts...)
	signature := signUrlWith(urlToSign, key, salt)
	return pathJoin(host, signature, urlToSign)
}

func getCanonicalUrl(assetKey string) string {
	s := "s3://" + assetKey
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

func signUrlWith(url string, key, salt string) string {
	keyBin, _ := hex.DecodeString(key)
	saltBin, _ := hex.DecodeString(salt)
	mac := hmac.New(sha256.New, keyBin)
	mac.Write(saltBin)
	mac.Write([]byte(url))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signature
}

func signUrl(url string) string {
	mac := hmac.New(sha256.New, config.Conf.ImgProxyAuth.GetKey())
	mac.Write(config.Conf.ImgProxyAuth.GetSalt())
	mac.Write([]byte(url))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signature
}
