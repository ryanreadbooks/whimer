package downloader

import (
	"context"
	"fmt"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
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

type Downloader struct {
	c   Config
	cli *minio.Client
}

func New(c Config) (*Downloader, error) {
	down := Downloader{
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
	down.cli = cli

	return &down, nil
}

func (d *Downloader) Download(ctx context.Context, bucket, object string) ([]byte, error) {
	obj, err := d.cli.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	content, err := io.ReadAll(obj)
	if err != nil {
		return nil, fmt.Errorf("downloader readAll: %w", err)
	}

	return content, nil
}

func (d *Downloader) DownloadReadCloser(ctx context.Context, bucket, object string) (io.ReadCloser, error) {
	obj, err := d.cli.GetObject(ctx, bucket, object, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (d *Downloader) DownloadImage(ctx context.Context, bucket, object string) (image.Image, error) {
	obj, err := d.DownloadReadCloser(ctx, bucket, object)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(obj)
	if err != nil {
		return nil, err
	}

	return img, nil
}
