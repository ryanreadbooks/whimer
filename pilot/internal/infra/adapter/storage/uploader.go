package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	miniocreds "github.com/minio/minio-go/v7/pkg/credentials"
	v7policy "github.com/minio/minio-go/v7/pkg/policy"
	v7set "github.com/minio/minio-go/v7/pkg/set"
	"github.com/ryanreadbooks/whimer/misc/oss/credentials"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/policy"
	"github.com/ryanreadbooks/whimer/misc/oss/policy/action"
	"github.com/ryanreadbooks/whimer/misc/oss/policy/condition"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xhttputil "github.com/ryanreadbooks/whimer/misc/xhttp/util"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xstring"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	modelerr "github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

const (
	postPolicyExpiration = time.Minute * 15
)

// Uploaders 上传器管理
type Uploaders struct {
	uploaders map[uploadresource.Type]*uploader
	ossConfig *config.Oss
}

// NewUploaders 创建上传器管理
func NewUploaders(c *config.Config, uploadOssCli *minio.Client) *Uploaders {
	u := &Uploaders{
		uploaders: make(map[uploadresource.Type]*uploader),
		ossConfig: &c.Oss,
	}
	for resourceType, metadata := range c.UploadResourceDefineMap {
		u.uploaders[resourceType] = newUploader(&c.UploadAuthSign, &c.Oss, uploadOssCli, resourceType, metadata)
	}
	return u
}

func (u *Uploaders) GetUploader(objType uploadresource.Type) (*uploader, error) {
	if uploader, ok := u.uploaders[objType]; ok {
		return uploader, nil
	}
	return nil, modelerr.ErrUnsupportedResource
}

func (u *Uploaders) SeperateObject(objType uploadresource.Type, fileId string) (bucket, key string, err error) {
	uploader, err := u.GetUploader(objType)
	if err != nil {
		return
	}
	bucket, key, ok := uploader.keyGen.Unwrap(fileId)
	if !ok {
		err = xerror.ErrArgs.Msg("对象格式错误")
		return
	}
	return
}

func (u *Uploaders) GetBucket(objType uploadresource.Type) (string, error) {
	uploader, err := u.GetUploader(objType)
	if err != nil {
		return "", err
	}
	return uploader.metadata.Bucket, nil
}

func (u *Uploaders) CheckFileIdValid(objType uploadresource.Type, fileId string) error {
	uploader, err := u.GetUploader(objType)
	if err != nil {
		return err
	}
	return uploader.CheckFileIdValid(fileId)
}

func (u *Uploaders) TrimBucketAndPrefix(objType uploadresource.Type, fileId string) string {
	uploader, err := u.GetUploader(objType)
	if err != nil {
		return fileId
	}
	return uploader.keyGen.TrimBucketAndPrefix(fileId)
}

// uploader 单个资源类型的上传器
type uploader struct {
	uploadSignConfig *config.UploadAuthSign
	ossConfig        *config.Oss
	ossCli           *minio.Client
	resourceType     uploadresource.Type
	metadata         uploadresource.Metadata

	keyGen      *keygen.Generator
	credentials *credentials.STSCredentials
	signer      *signer.JwtUploadAuthSigner
}

func newUploader(
	c *config.UploadAuthSign,
	ossConfig *config.Oss,
	ossCli *minio.Client,
	resource uploadresource.Type, metadata uploadresource.Metadata,
) *uploader {
	u := &uploader{
		uploadSignConfig: c,
		ossConfig:        ossConfig,
		ossCli:           ossCli,
		resourceType:     resource,
		metadata:         metadata,
	}

	u.keyGen = keygen.NewGenerator(
		keygen.WithBucket(metadata.Bucket),
		keygen.WithPrefix(metadata.Prefix),
		keygen.WithPrependBucket(metadata.PrependBucket),
		keygen.WithPrependPrefix(metadata.PrependPrefix),
		keygen.WithStringer(metadata.GetStringer()),
	)

	u.signer = signer.NewJwtUploadAuthSigner(&signer.JwtSignConfig{
		JwtIssuer:   c.JwtIssuer,
		JwtSubject:  c.JwtSubject,
		JwtDuration: c.JwtDuration,
		Ak:          xstring.AsBytes(c.Ak),
		Sk:          xstring.AsBytes(c.Sk),
	})

	// make upload policy
	keyPrefix := ""
	if metadata.Prefix != "" && metadata.PrependPrefix {
		keyPrefix = fmt.Sprintf("%s/%s/*", metadata.Bucket, metadata.Prefix)
	} else {
		keyPrefix = fmt.Sprintf("%s/*", metadata.Bucket)
	}
	commonConditionsMap := make(v7policy.ConditionKeyMap)
	commonConditionsMap.Add(condition.S3SignatureVersion, v7set.CreateStringSet(condition.SignatureV4))

	p := policy.New()
	stmt := policy.NewAllowStatement()
	stmt.Actions.Add(action.PutObject)
	stmt.Actions.Add(action.ListMultipartUploadParts)
	stmt.Actions.Add(action.AbortMultipartUpload)
	stmt.Resources.Add(policy.GetSimpleResource(keyPrefix))
	stmt.Conditions.Add(condition.StringEquals, commonConditionsMap)

	p.AppendStatement(stmt)

	xlog.Msgf("%s policy is %s", keyPrefix, p.String()).Info()

	endpoint := xhttputil.FormatHost(ossConfig.Endpoint, ossConfig.UseSecure)
	creds, err := credentials.NewSTSCredentials(credentials.Config{
		Endpoint:        endpoint,
		AccessKey:       ossConfig.User,
		SecretKey:       ossConfig.Password,
		DurationSeconds: ossConfig.CredentialDurationSec,
		Policy:          p.String(),
	})
	if err != nil {
		panic(err)
	}
	u.credentials = creds
	_, err = u.credentials.Get(context.Background())
	if err != nil {
		panic(err)
	}

	return u
}

func (u *uploader) GetFileIds(cnt int32) []string {
	fileIds := make([]string, 0, cnt)
	for range cnt {
		fileIds = append(fileIds, u.keyGen.Gen())
	}
	return fileIds
}

func (u *uploader) CheckFileIdValid(fileId string) error {
	if !u.keyGen.Check(fileId) {
		return xerror.ErrArgs.Msg("资源格式错误")
	}
	return nil
}

func (u *uploader) GetBucket() string {
	return u.metadata.Bucket
}

func (u *uploader) GetMetadata() uploadresource.Metadata {
	return u.metadata
}

func (u *uploader) GetUploadEndpoint() string {
	return xhttputil.FormatHost(u.ossConfig.UploadEndpoint, u.ossConfig.UseSecure)
}

func (u *uploader) GetCredentials(ctx context.Context) (*miniocreds.Value, error) {
	val, err := u.credentials.Get(ctx)
	return &val, err
}

type UploadTicket struct {
	FileIds     []string `json:"file_ids"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}

// Should be deprecated in the future
func (u *uploader) GenerateUploadTicket(count int32, source string) (*UploadTicket, error) {
	fileIds := make([]string, 0, count)
	for range count {
		fileIds = append(fileIds, u.keyGen.Gen())
	}

	res, err := u.signer.BatchGetUploadAuth(fileIds, string(u.resourceType))
	if err != nil {
		xlog.Msg("signer batch get upload auth failed").Err(err).Error()
		return nil, modelerr.ErrServerSignFailure
	}

	return &UploadTicket{
		FileIds:     fileIds,
		CurrentTime: res.CurrentTime,
		ExpireTime:  res.ExpireTime,
		Token:       res.Token,
		UploadAddr:  u.ossConfig.UploadEndpoint,
	}, nil
}

// GetPostPolicyRequest Post Policy 请求参数
type GetPostPolicyRequest struct {
	ContentType string
	Sha256      string
}

// GetPostPolicyResponse Post Policy 响应
type GetPostPolicyResponse struct {
	Url  string
	Key  string
	Form map[string]string
}

func (u *uploader) GetPostPolicy(ctx context.Context, req *GetPostPolicyRequest) (*GetPostPolicyResponse, error) {
	p := minio.NewPostPolicy()
	key := u.keyGen.Gen()
	p.SetBucket(u.metadata.Bucket)
	_, rawKey, ok := u.keyGen.Unwrap(key)
	if !ok {
		return nil, xerror.ErrInternal.Msg("key format error")
	}
	p.SetKey(rawKey)
	p.SetContentType(req.ContentType)
	p.SetExpires(time.Now().Add(postPolicyExpiration))
	p.SetContentLengthRange(1, u.resourceType.PermitSize())
	p.SetChecksum(minio.NewChecksumString(minio.ChecksumSHA256, req.Sha256))

	url, form, err := u.ossCli.PresignedPostPolicy(ctx, p)
	if err != nil {
		return nil, err
	}

	return &GetPostPolicyResponse{
		Url:  url.String(),
		Key:  key,
		Form: form,
	}, nil
}
