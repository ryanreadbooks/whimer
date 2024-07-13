package uploader

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestMinioUpload(t *testing.T) {
	up, err := New(Config{
		Ak:       os.Getenv("ENV_OSS_AK"),
		Sk:       os.Getenv("ENV_OSS_SK"),
		Endpoint: "127.0.0.1:9000",
		Location: "local",
	})
	if err != nil {
		t.Fatal(err)
	}

	buf := []byte(`this is a hello world text made for minio test`)

	err = up.Upload(context.Background(), &UploadMeta{
		Bucket:      "test-bucket",
		Name:        "uploader-test-1.txt",
		Buf:         buf,
		ContentType: "text/plain",
	})
	t.Log(err)

	// 预签名
	url, err := up.cli.PresignedGetObject(context.TODO(), "test-bucket", "uploader-test-1.txt", time.Minute, nil)
	t.Log(err)
	t.Log(url)

	// 删除
	err = up.Remove(context.Background(), "test-bucket", "uploader-test-1.txt")
	t.Log(err)
}
