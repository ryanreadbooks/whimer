package uploader

import (
	"testing"

	"github.com/minio/minio-go/v7/pkg/s3utils"
)

func TestEncodePath(t *testing.T) {
	p := s3utils.EncodePath("/nota-prv/2f5ae251edddc8319998c75c3ea0323d/rrv4@prv_webp_50")
	t.Log(p)
}