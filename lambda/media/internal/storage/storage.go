package storage

import (
	"context"
	"io"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string `json:"endpoint"`
	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`
	UseSSL    bool   `json:"useSsl"`
}

type Storage struct {
	client *minio.Client
}

func New(c Config) (*Storage, error) {
	client, err := minio.New(c.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.AccessKey, c.SecretKey, ""),
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

func (s *Storage) UploadFromOutput(
	ctx context.Context,
	bucket, key string,
	filePath string,
	reader io.ReadCloser,
	contentType string,
) error {
	if filePath != "" {
		defer os.Remove(filePath)
		return s.UploadFile(ctx, bucket, key, filePath, contentType)
	}
	if reader != nil {
		defer reader.Close()
		return s.UploadStream(ctx, bucket, key, reader, contentType)
	}
	return nil
}
