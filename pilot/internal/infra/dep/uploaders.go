package dep

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

var uploaders *Uploaders

type Uploaders struct {
	uploaders        map[uploadresource.Type]*uploader
	displayOssClient *minio.Client
	uploadOssClient  *minio.Client
	ossConfig        *config.Oss
}

func initUploaders(c *config.Config) {
	uploaders = &Uploaders{
		uploaders:        make(map[uploadresource.Type]*uploader),
		displayOssClient: displayOssCli,
		uploadOssClient:  uploadOssCli,
		ossConfig:        &c.Oss,
	}
	for resourceType, metadata := range c.UploadResourceDefineMap {
		uploaders.uploaders[resourceType] = newUploader(&c.UploadAuthSign, &c.Oss, uploadOssCli, resourceType, metadata)
	}
}

func GetUploaders() *Uploaders {
	return uploaders
}

func (u *Uploaders) GetUploader(resource uploadresource.Type) (*uploader, error) {
	if uploader, ok := u.uploaders[resource]; ok {
		return uploader, nil
	}
	return nil, modelerr.ErrUnsupportedResource
}

// 对外返回预签名url获取资源
func (u *Uploaders) PresignGetUrl(ctx context.Context, resource uploadresource.Type, key string) (string, error) {
	uploader, err := u.GetUploader(resource)
	if err != nil {
		return "", err
	}
	bucket := uploader.metadata.Bucket
	_, rawKey, ok := uploader.keyGen.Unwrap(key)
	if !ok {
		return "", xerror.ErrArgs.Msg("资源格式错误")
	}

	presignedURL, err := u.displayOssClient.PresignedGetObject(ctx, bucket, rawKey, time.Hour, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func (u *Uploaders) SeperateResource(resource uploadresource.Type, resourceId string) (bucket, key string, err error) {
	uploader, err := u.GetUploader(resource)
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

func (u *Uploaders) GetBucket(resource uploadresource.Type) (string, error) {
	uploader, err := u.GetUploader(resource)
	if err != nil {
		return "", err
	}
	return uploader.metadata.Bucket, nil
}

func (u *Uploaders) CheckFileIdValid(resource uploadresource.Type, fileId string) error {
	uploader, err := u.GetUploader(resource)
	if err != nil {
		return err
	}
	return uploader.CheckFileIdValid(fileId)
}

// 去除 bucket 和 prefix
func (u *Uploaders) TrimBucketAndPrefix(resource uploadresource.Type, fileId string) string {
	uploader, err := u.GetUploader(resource)
	if err != nil {
		return fileId
	}
	return uploader.keyGen.TrimBucketAndPrefix(fileId)
}

func PresignGetUrl(ctx context.Context, resource uploadresource.Type, key string) (string, error) {
	return uploaders.PresignGetUrl(ctx, resource, key)
}

func SeperateResource(resource uploadresource.Type, resourceId string) (bucket, key string, err error) {
	return uploaders.SeperateResource(resource, resourceId)
}

func GetBucket(resource uploadresource.Type) (string, error) {
	return uploaders.GetBucket(resource)
}

type uploader struct {
	uploadSignConfig *config.UploadAuthSign
	ossConfig        *config.Oss
	ossCli           *minio.Client
	resourceType     uploadresource.Type // 当前uploader只负责一种资源的上传凭证申请
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
	stmt.Resources.Add(policy.GetSimpleResource(keyPrefix)) // 只能上传到特定的桶和前缀
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

func (u *uploader) GetUploadEndpoint() string {
	return xhttputil.FormatHost(u.ossConfig.UploadEndpoint, u.ossConfig.UseSecure)
}

func (u *uploader) GetCredentials(ctx context.Context) (*miniocreds.Value, error) {
	val, err := u.credentials.Get(ctx)
	return &val, err
}

type GetPostPolicyRequest struct {
	ContentType string
	Sha256      string
}

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
