package dto

import (
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

type NoteImageMetadata struct {
	Format string `json:"format"`
	Height uint32 `json:"height"`
	Width  uint32 `json:"width"`
}

type NoteImage struct {
	Key        string            `json:"key,omitempty"`
	Url        string            `json:"url"`
	UrlPreview string            `json:"url_prv"`
	Type       notevo.AssetType  `json:"type"`
	Metadata   NoteImageMetadata `json:"metadata"`
}

func NewNoteImageUrl(key string) string {
	bucket := storagevo.ObjectTypeNoteImage.Metadata().Bucket
	url := imgproxy.GetSignedUrl(
		config.Conf.Oss.DisplayEndpointBucket(bucket),
		key,
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality(config.Conf.ImgQuality.Quality),
	)
	return url
}

func ConvertEntityNoteImageToDto(image *entity.NoteImage) *NoteImage {
	if image == nil {
		return nil
	}

	return &NoteImage{
		Key:        image.FileId,
		Url:        NewNoteImageUrl(image.FileId),
		UrlPreview: NewNoteImagePreviewUrl(image.FileId),
		Type:       notevo.AssetTypeImage,
		Metadata: NoteImageMetadata{
			Width:  image.Width,
			Height: image.Height,
			Format: image.Format,
		},
	}
}

func NewNoteImagePreviewUrl(key string) string {
	bucket := storagevo.ObjectTypeNoteImage.Metadata().Bucket
	url := imgproxy.GetSignedUrl(
		config.Conf.Oss.DisplayEndpointBucket(bucket),
		key,
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality(config.Conf.ImgQuality.QualityPreview))
	return url
}

type NoteImageList []*NoteImage

type NoteVideoMeta struct {
	Width      uint32  `json:"width"`
	Height     uint32  `json:"height"`
	Format     string  `json:"format"`
	Duration   float64 `json:"duration"`
	Bitrate    int64   `json:"bitrate"`
	Codec      string  `json:"codec"`
	Framerate  float64 `json:"framerate"`
	AudioCodec string  `json:"audio_codec"`
}

type NoteVideo struct {
	Key      string           `json:"-"`
	Url      string           `json:"url"`
	Type     notevo.AssetType `json:"type"`
	Metadata NoteVideoMeta    `json:"metadata"`
}

func ConvertEntityNoteVideoToDto(video *entity.NoteVideo, url string) *NoteVideo {
	if video == nil {
		return nil
	}

	return &NoteVideo{
		Key:  video.FileId,
		Url:  url,
		Type: notevo.AssetTypeVideo,
		Metadata: NoteVideoMeta{
			Width:      video.GetMetadata().Width,
			Height:     video.GetMetadata().Height,
			Format:     video.GetMetadata().Format,
			Duration:   video.GetMetadata().Duration,
			Bitrate:    video.GetMetadata().Bitrate,
			Codec:      video.GetMetadata().Codec,
			Framerate:  video.GetMetadata().Framerate,
			AudioCodec: video.GetMetadata().AudioCodec,
		},
	}
}

type NoteVideoList []*NoteVideo

type NoteTag struct {
	Id   notevo.TagId `json:"id"`
	Name string       `json:"name"`
}

type NoteInteraction struct {
	Liked     bool `json:"liked"`     // 用户是否点赞过该笔记
	Commented bool `json:"commented"` // 用户是否评论过该笔记
}
