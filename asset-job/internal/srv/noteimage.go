package srv

import (
	"bytes"
	"context"
	"strings"

	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/asset-job/internal/infra/oss"
	"github.com/ryanreadbooks/whimer/asset-job/internal/model"
	"github.com/ryanreadbooks/whimer/misc/oss/uploader"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"github.com/chai2010/webp"
	"github.com/minio/minio-go/v7/pkg/notification"
)

type NoteImageService struct {
}

func NewNoteImageService() *NoteImageService {
	s := NoteImageService{}

	return &s
}

const previewSuffix = "_prv_webp_50"

// 笔记图片上传成功
func (s *NoteImageService) OnImageUploaded(ctx context.Context, event *model.MinioEvent) error {
	if event.EventName != string(notification.ObjectCreatedPut) {
		return nil
	}

	bucket := config.Conf.NoteOss.Bucket
	// event.Key is represented as bucket + / + pureKey
	pureKey := strings.TrimLeft(event.Key, bucket+"/")
	image, err := oss.Downloader().DownloadImage(ctx, bucket, pureKey)
	if err != nil {
		return xerror.Wrapf(err, "OnImageUploaded download image failed")
	}

	// do something with image
	x, y := image.Bounds().Max.X, image.Bounds().Max.Y
	buf := make([]byte, 0, x*y)
	var writer = bytes.NewBuffer(buf)
	err = webp.Encode(writer, image, &webp.Options{
		Lossless: false,
		Quality:  50,
		Exact:    true,
	})
	if err != nil {
		return xerror.Wrapf(err, "OnImageUploaded webp Encode failed")
	}

	// upload back to oss again
	err = oss.Uploader().Upload(ctx, &uploader.UploadMeta{
		Bucket:      config.Conf.NoteOss.PrvBucket,
		Name:        pureKey + previewSuffix,
		Buf:         writer.Bytes(),
		ContentType: "image/webp",
	})
	if err != nil {
		return xerror.Wrapf(err, "OnImageUploaded failed to upload preview webp image to oss")
	}

	xlog.Msg("OnImageUploaded handled ok").Extras("bucket", bucket, "key", pureKey, "x", x, "y", y).Infox(ctx)

	return nil
}
