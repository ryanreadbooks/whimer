package upload

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xstring"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

type uploader struct {
	c            *config.UploadAuthSign
	oss          *config.Oss
	resourceType uploadresource.Type // 配置类型来自配置文件
	metadata     uploadresource.Metadata

	keyGen *keygen.Generator
	signer *signer.JwtUploadAuthSigner
}

func newUploader(c *config.UploadAuthSign, oss *config.Oss,
	resource uploadresource.Type, metadata uploadresource.Metadata) *uploader {
	u := &uploader{
		c:            c,
		oss:          oss,
		resourceType: resource,
		metadata:     metadata,
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

	return u
}

type StsTicket struct {
	FileIds     []string `json:"file_ids"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}

func (u *uploader) generateStsTicket(count int32, source string) (*StsTicket, error) {
	fileIds := make([]string, 0, count)
	for range count {
		fileIds = append(fileIds, u.keyGen.Gen())
	}

	res, err := u.signer.BatchGetUploadAuth(fileIds, string(u.resourceType))
	if err != nil {
		xlog.Msg("signer batch get upload auth failed").Err(err).Error()
		return nil, errors.ErrServerSignFailure
	}

	return &StsTicket{
		FileIds:     fileIds,
		CurrentTime: res.CurrentTime,
		ExpireTime:  res.ExpireTime,
		Token:       res.Token,
		UploadAddr:  u.oss.UploadEndpoint,
	}, nil
}
