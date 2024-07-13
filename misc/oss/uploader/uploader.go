package uploader

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Ak       string
	Sk       string
	Endpoint string
	Location string
	Secure   bool
}

// 使用ak/sk进行上传
// 一般是服务端侧代理上传操作会用到这个
type Uploader struct {
	c   Config
	cli *minio.Client
}

type UploadMeta struct {
	Bucket      string
	Name        string
	Content     io.Reader
	Buf         []byte // 如果和Content同时存在的话，优先使用Content
	ContentType string
}

func New(c Config) (*Uploader, error) {
	up := &Uploader{
		c: c,
	}

	cli, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Ak, c.Sk, ""),
		Region: c.Location,
		Secure: c.Secure,
	})
	if err != nil {
		return nil, err
	}

	up.cli = cli

	return up, nil
}

func (u *Uploader) Upload(ctx context.Context, obj *UploadMeta) error {
	if obj == nil {
		return errors.New("obj is empty")
	}

	var size int64
	if obj.Content == nil {
		obj.Content = bytes.NewBuffer(obj.Buf)
		size = int64(len(obj.Buf))
	}

	_, err := u.cli.PutObject(ctx,
		obj.Bucket,
		obj.Name,
		obj.Content,
		size,
		minio.PutObjectOptions{
			ContentType: obj.ContentType,
		})
	if err != nil {
		return err
	}

	return nil
}

func (u *Uploader) Remove(ctx context.Context, bucket, objectName string) error {
	return u.cli.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (u *Uploader) GetPublicVisitUrl(bucket, objectName, replaceEndpoint string) string {
	endpoint := u.c.Endpoint
	if replaceEndpoint != "" {
		endpoint = replaceEndpoint
	}
	return fmt.Sprintf("%s/%s/%s", endpoint, bucket, objectName)
}
