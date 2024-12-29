package signer

import (
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/minio/minio-go/v7/pkg/credentials"
	miniosigner "github.com/minio/minio-go/v7/pkg/signer"
)

// 使用临时用户生成签名的上传凭证
type Signer struct {
	sync.Mutex
	c        Config
	user     string
	password string

	cred *credentials.Credentials
}

type Config struct {
	Endpoint string
	Location string
}

type SignInfo struct {
	Auth     string
	Sha256   string
	Date     string
	Token    string
	ExpireAt time.Time
}

func NewSigner(user, pass string, c Config) *Signer {
	// 初始化一些可用的credentials
	ep := c.Endpoint
	epWithSchema := ep
	if !strings.HasPrefix(ep, "http://") && !strings.HasPrefix(ep, "https://") {
		epWithSchema = "http://" + ep
	}
	cred, _ := credentials.NewSTSAssumeRole(epWithSchema,
		credentials.STSAssumeRoleOptions{
			AccessKey:       user,
			SecretKey:       pass,
			Location:        c.Location,
			DurationSeconds: int(time.Hour.Seconds()),
		})

	c.Endpoint = strings.TrimPrefix(c.Endpoint, "http://")
	c.Endpoint = strings.TrimPrefix(c.Endpoint, "https://")
	s := &Signer{
		c:    c,
		cred: cred,
	}

	return s
}

func (s *Signer) Sign(path string) (*SignInfo, error) {
	req := http.Request{
		Header: make(http.Header, 0),
		Method: http.MethodPut,
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	req.URL = &url.URL{
		Host: s.c.Endpoint,
		Path: path,
	}
	// 不签名body
	req.Header.Add("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")

	val, err := s.cred.Get()
	if err != nil {
		return nil, err
	}

	// 获取签名所需凭证
	nReq := miniosigner.SignV4(req, val.AccessKeyID, val.SecretAccessKey, val.SessionToken, s.c.Location)

	return &SignInfo{
		Auth:     nReq.Header.Get("Authorization"),
		Date:     nReq.Header.Get("X-Amz-Date"),
		Sha256:   nReq.Header.Get("X-Amz-Content-Sha256"),
		Token:    nReq.Header.Get("X-Amz-Security-Token"),
		ExpireAt: val.Expiration,
	}, nil
}
