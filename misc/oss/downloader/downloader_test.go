package downloader

import (
	"context"
	"os"
	"testing"
)

var down *Downloader

func TestMain(m *testing.M) {
	var err error
	down, err = New(Config{
		Ak:       os.Getenv("ENV_OSS_AK"),
		Sk:       os.Getenv("ENV_OSS_SK"),
		Endpoint: "127.0.0.1:9000",
		Location: "local",
	})
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestMinioDownload(t *testing.T) {
	content, err := down.Download(context.Background(), "nota", "41e08695953e6bc96ee1e7ef1b9a11538c345bc8")
	if err != nil {
		t.Error(err)
	}

	if len(content) == 0 {
		t.Error("content should not be zero-length")
	}
}

func TestMinioDownloadImage(t *testing.T) {
	img, err := down.DownloadImage(context.TODO(), "nota", "105e92df3051fcb5ca1701c6d7e448fe5b2a083f")
	if err != nil {
		t.Error(err)
	}

	t.Log(img.Bounds().Max)
}
