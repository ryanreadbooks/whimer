package storage

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

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
)

type uploader struct {
	uploadSignConfig *config.UploadAuthSign
	oss              *config.Oss
	resourceType     uploadresource.Type // 配置类型来自配置文件
	metadata         uploadresource.Metadata

	keyGen      *keygen.Generator
	credentials *credentials.STSCredentials
	signer      *signer.JwtUploadAuthSigner
}

func newUploader(c *config.UploadAuthSign, oss *config.Oss,
	resource uploadresource.Type, metadata uploadresource.Metadata) *uploader {
	u := &uploader{
		uploadSignConfig: c,
		oss:              oss,
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

	endpoint := xhttputil.FormatHost(oss.Endpoint, oss.UseSecure)
	creds, err := credentials.NewSTSCredentials(credentials.Config{
		Endpoint:        endpoint,
		AccessKey:       oss.User,
		SecretKey:       oss.Password,
		DurationSeconds: oss.CredentialDurationSec,
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

func (u *uploader) generateUploadTicket(count int32, source string) (*UploadTicket, error) {
	fileIds := make([]string, 0, count)
	for range count {
		fileIds = append(fileIds, u.keyGen.Gen())
	}

	res, err := u.signer.BatchGetUploadAuth(fileIds, string(u.resourceType))
	if err != nil {
		xlog.Msg("signer batch get upload auth failed").Err(err).Error()
		return nil, errors.ErrServerSignFailure
	}

	return &UploadTicket{
		FileIds:     fileIds,
		CurrentTime: res.CurrentTime,
		ExpireTime:  res.ExpireTime,
		Token:       res.Token,
		UploadAddr:  u.oss.UploadEndpoint,
	}, nil
}

func (u *uploader) getFileIds(cnt int32) []string {
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
