package dep

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	xhttpctransport "github.com/ryanreadbooks/whimer/misc/xhttp/client/transport"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	ossCli        *minio.Client
	displayOssCli *minio.Client
)

// InitOss 初始化对象存储客户端
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

	displayCli, err := minio.New(
		strings.TrimPrefix(strings.TrimPrefix(c.Oss.DisplayEndpoint, "https://"), "http://"),
		&minio.Options{
			Secure:    c.Oss.UseSecure,
			Creds:     credentials.NewStaticV4(c.Oss.User, c.Oss.Password, ""),
			Transport: xhttpctransport.SpanTracing(http.DefaultTransport),
		})
	if err != nil {
		panic(fmt.Errorf("init display oss: %w", err))
	}

	ossCli = cli
	displayOssCli = displayCli
}

func OssClient() *minio.Client {
	return ossCli
}
