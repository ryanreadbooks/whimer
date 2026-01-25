package storage

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xhttputil "github.com/ryanreadbooks/whimer/misc/xhttp/util"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
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

func (b *Biz) SeperateResource(resource uploadresource.Type, resourceId string) (bucket, key string, err error) {
	return b.uploaders.SeperateResource(resource, resourceId)
}

// Deprecated
func (b *Biz) RequestUploadTicket(ctx context.Context,
	resource uploadresource.Type, cnt int32, source string,
) (*dep.UploadTicket, error) {
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

type GetUploadTemporaryTicketRequest struct {
	Resource uploadresource.Type
	Count    int32
	Source   string
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

// 获取STS临时上传凭证
func (b *Biz) GetUploadTemporaryTicket(
	ctx context.Context,
	req *GetUploadTemporaryTicketRequest,
) (*TemporaryCredentials, error) {
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

type GetPostPolicyUploadTicketRequest struct {
	Resource uploadresource.Type
	Sha256   string
	Size     int64
	MimeType string
}

type GetPostPolicyUploadTicketResponse struct {
	FileId     string            `json:"file_id"`
	UploadAddr string            `json:"upload_addr"`
	Form       map[string]string `json:"form"`
}

// 获取post policy临时上传凭证
func (b *Biz) GetPostPolicyUploadTicket(
	ctx context.Context,
	req *GetPostPolicyUploadTicketRequest,
) (*GetPostPolicyUploadTicketResponse, error) {
	uploader, err := b.uploaders.GetUploader(req.Resource)
	if err != nil {
		return nil, err
	}

	ppResp, err := uploader.GetPostPolicy(ctx, &dep.GetPostPolicyRequest{
		ContentType: req.MimeType,
		Sha256:      req.Sha256,
	})
	if err != nil {
		return nil, err
	}

	// Post Policy 上传需要 POST 到 bucket 的根路径，key 通过 form 字段传递
	return &GetPostPolicyUploadTicketResponse{
		UploadAddr: ppResp.Url,
		FileId:     ppResp.Key,
		Form:       ppResp.Form,
	}, nil
}
