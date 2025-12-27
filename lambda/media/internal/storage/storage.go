package storage

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint string `json:"endpoint"`
	Ak       string `json:"ak"`
	Sk       string `json:"sk"`
	UseSSL   bool   `json:"useSsl"`
}

type Storage struct {
	client *minio.Client
}

func New(c Config) (*Storage, error) {
	client, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.Ak, c.Sk, ""),
		Secure: c.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Storage{client: client}, nil
}

func (s *Storage) UploadFile(ctx context.Context, bucket, key, filePath string, contentType string) error {
	_, err := s.client.FPutObject(ctx, bucket, key, filePath, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) UploadStream(ctx context.Context, bucket, key string, reader io.Reader, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, key, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Download(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
}

func (s *Storage) DownloadToFile(ctx context.Context, bucket, key, filePath string) error {
	return s.client.FGetObject(ctx, bucket, key, filePath, minio.GetObjectOptions{})
}

func (s *Storage) Delete(ctx context.Context, bucket, key string) error {
	return s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{})
}

func (s *Storage) GetObjectURL(bucket, key string) string {
	return s.client.EndpointURL().String() + "/" + bucket + "/" + key
}

// GetPresignedURL 生成签名 URL，用于临时访问私有对象
func (s *Storage) GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	presignedURL, err := s.client.PresignedGetObject(ctx, bucket, key, expires, url.Values{})
	if err != nil {
		return "", err
	}
	return presignedURL.String(), nil
}
