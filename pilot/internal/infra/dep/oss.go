package dep

import (
	"fmt"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	xhttpctransport "github.com/ryanreadbooks/whimer/misc/xhttp/client/transport"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	ossCli *minio.Client
)

// 对象存储
func InitOss(c *config.Config) {
	cli, err := minio.New(c.Oss.Endpoint,
		&minio.Options{
			Secure:    c.Oss.UseSecure,
			Creds:     credentials.NewStaticV4(c.Oss.User, c.Oss.Password, ""),
			Transport: xhttpctransport.SpanTracing(http.DefaultTransport),
		})
	if err != nil {
		panic(fmt.Errorf("init oss: %w", err))
	}

	ossCli = cli
}

func OssClient() *minio.Client {
	return ossCli
}
