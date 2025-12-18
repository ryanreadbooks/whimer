package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xhttputil "github.com/ryanreadbooks/whimer/misc/xhttp/util"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

type Biz struct {
	uploaders *dep.Uploaders
	ossClient *minio.Client
}

func NewBiz(c *config.Config) *Biz {
	return &Biz{
		uploaders: dep.GetUploaders(),
		ossClient: dep.OssClient(),
	}
}

func (b *Biz) PresignGetUrl(ctx context.Context, resource uploadresource.Type, key string) (string, error) {
	return b.uploaders.PresignGetUrl(ctx, resource, key)
}

func (b *Biz) SeperateResource(resource uploadresource.Type, resourceId string) (bucket, key string, err error) {
	return b.uploaders.SeperateResource(resource, resourceId)
}

// Deprecated
func (b *Biz) RequestUploadTicket(ctx context.Context, resource uploadresource.Type, cnt int32, source string) (*dep.UploadTicket, error) {
	uploader, err := b.uploaders.GetUploader(resource)
	if err != nil {
		return nil, err
	}

	ticket, err := uploader.GenerateUploadTicket(cnt, source)
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader generate ticket failed").WithCtx(ctx)
	}

	return ticket, nil
}

type RequestUploadTemporaryTicket struct {
	Resource uploadresource.Type
	Count    int32
	Source   string
}

// 获取STS临时上传凭证
func (b *Biz) RequestUploadTemporaryTicket(ctx context.Context, req RequestUploadTemporaryTicket) (*TemporaryCredentials, error) {
	uploader, err := b.uploaders.GetUploader(req.Resource)
	if err != nil {
		return nil, err
	}

	tmpCreds, err := uploader.GetCredentials(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader %s failed to get sts credentials", req.Resource).WithCtx(ctx)
	}

	return &TemporaryCredentials{
		FileIds:      uploader.GetFileIds(req.Count),
		Bucket:       uploader.GetBucket(),
		AccessKey:    tmpCreds.AccessKeyID,
		SecretKey:    tmpCreds.SecretAccessKey,
		SessionToken: tmpCreds.SessionToken,
		ExpireAt:     tmpCreds.Expiration.Unix(),
		UploadAddr:   xhttputil.FormatHost(uploader.GetUploadEndpoint(), false),
	}, nil
}

type TemporaryCredentials struct {
	FileIds      []string `json:"-"`
	Bucket       string   `json:"-"`
	AccessKey    string   `json:"tmp_access_key"`
	SecretKey    string   `json:"tmp_secret_key"`
	SessionToken string   `json:"session_token"`
	ExpireAt     int64    `json:"expire_at"` // unix timestamp in second
	UploadAddr   string   `json:"upload_addr"`
}
