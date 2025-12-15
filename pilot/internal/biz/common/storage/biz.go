package storage

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	xhttputil "github.com/ryanreadbooks/whimer/misc/xhttp/util"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

	"github.com/minio/minio-go/v7"
)

type Biz struct {
	resourceDefine config.UploadResourceDefineMap
	uploaders      map[uploadresource.Type]*uploader
	ossClient      *minio.Client
}

func NewBiz(c *config.Config) *Biz {
	b := &Biz{
		resourceDefine: c.UploadResourceDefineMap,
		ossClient:      dep.OssClient(),
	}

	b.uploaders = make(map[uploadresource.Type]*uploader)
	for resourceType, metadata := range b.resourceDefine {
		b.uploaders[resourceType] = newUploader(&c.UploadAuthSign, &c.Oss, resourceType, metadata)
	}

	return b
}

func (b *Biz) getUploader(resource uploadresource.Type) (*uploader, error) {
	if uploader, ok := b.uploaders[resource]; ok {
		return uploader, nil
	}

	return nil, modelerr.ErrUnsupportedResource
}


func (b *Biz) SeperateResource(resource uploadresource.Type, resourceId string) (bucket, key string, err error) {
	uploader, err := b.getUploader(resource)
	if err != nil {
		return
	}
	bucket, key, ok := uploader.keyGen.Unwrap(resourceId)
	if !ok {
		err = xerror.ErrArgs.Msg("资源格式错误")
		return
	}

	return
}

// Deprecated
func (b *Biz) RequestUploadTicket(ctx context.Context, resource uploadresource.Type, cnt int32, source string) (*UploadTicket, error) {
	uploader, err := b.getUploader(resource)
	if err != nil {
		return nil, err
	}

	ticket, err := uploader.generateUploadTicket(cnt, source)
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
	uploader, err := b.getUploader(req.Resource)
	if err != nil {
		return nil, err
	}

	tmpCreds, err := uploader.credentials.Get(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader %s failed to get sts credentials", req.Resource).WithCtx(ctx)
	}

	return &TemporaryCredentials{
		FileIds:      uploader.getFileIds(req.Count),
		Bucket:       uploader.metadata.Bucket,
		AccessKey:    tmpCreds.AccessKeyID,
		SecretKey:    tmpCreds.SecretAccessKey,
		SessionToken: tmpCreds.SessionToken,
		ExpireAt:     tmpCreds.Expiration.Unix(),
		UploadAddr:   xhttputil.FormatHost(uploader.oss.UploadEndpoint, uploader.oss.UseSecure),
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
