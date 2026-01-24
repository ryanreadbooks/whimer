package assetprocess

import (
	"context"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	conductor "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type VideoProcessor struct {
	baseProcessor

	biz *biz.Biz
}

func newVideoProcessor(biz *biz.Biz) Processor {
	return &VideoProcessor{biz: biz}
}

// 视频处理worker支持的请求参数
type VideoProcessReqParam struct {
	InputMode   string              `json:"input_mode"`
	InputBucket string              `json:"input_bucket"`
	InputKey    string              `json:"input_key"`
	Outputs     []*VideoProcessDest `json:"outputs"`
}

type VideoProcessDest struct {
	Bucket    string          `json:"bucket"`
	OutputKey string          `json:"output_key"`
	Settings  *EncodeSettings `json:"settings"`
}

type EncodeSettings struct {
	VideoCodec   string `json:"video_codec,omitempty"`
	AudioCodec   string `json:"audio_codec,omitempty"`
	VideoBitrate string `json:"video_bitrate,omitempty"`
	AudioBitrate string `json:"audio_bitrate,omitempty"`
	MaxHeight    int    `json:"max_height,omitempty"`
	MaxWidth     int    `json:"max_width,omitempty"`
	Preset       string `json:"preset,omitempty"`
	CRF          int    `json:"crf,omitempty"`
}

func getVideoOutputAssets(note *model.Note) []*VideoProcessDest {
	// key中包含了bucket 这里要将bucket去掉

	assets := make([]*VideoProcessDest, 0, 3)
	for _, item := range note.Videos.Items {
		codec := "libx264"
		preset := "medium"
		crf := 26
		maxHeight := 1080
		if strings.HasSuffix(item.Key, "_265.mp4") {
			codec = "libx265"
			preset = "medium"
			crf = 28
			maxHeight = 1080
		} else if strings.HasSuffix(item.Key, "_av1.mp4") {
			codec = "libsvtav1"
			preset = "8"
			crf = 35
			maxHeight = 1080
		}
		assets = append(assets, &VideoProcessDest{
			Bucket:    item.GetBucket(),
			OutputKey: item.TrimBucket(),
			Settings: &EncodeSettings{
				VideoCodec: codec,
				Preset:     preset,
				CRF:        crf,
				MaxHeight:  maxHeight,
			},
		})
	}

	return assets
}

func (p *VideoProcessor) Process(ctx context.Context, note *model.Note) (string, error) {
	param := FormatVideoProcessParam(note)
	// 调用视频处理任务
	callbackUrl := encodeCallbackUrl(
		config.Conf.DevCallbacks.NoteProcessCallback,
		note.NoteId,
		map[string]string{
			"asset_type": "video",
		})
	taskId, err := dep.GetConductProducer().Schedule(
		ctx,
		global.NoteVideoProcessTaskType,
		param,
		conductor.ScheduleOptions{
			Namespace:   global.NoteProcessNamespace,
			CallbackUrl: callbackUrl,
			MaxRetry:    3,
			ExpireAfter: time.Hour * 2,
		},
	)
	if err != nil {
		return "", xerror.Wrapf(err, "srv creator schedule task failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return taskId, nil
}

func FormatVideoProcessParam(note *model.Note) *VideoProcessReqParam {
	rawFileId := note.Videos.GetRawUrl()
	idx := strings.Index(rawFileId, "/")
	rawBucket := rawFileId[:idx]
	inputKey := strings.TrimPrefix(rawFileId, rawBucket+"/")

	param := &VideoProcessReqParam{
		InputMode:   "s3",
		InputBucket: rawBucket,
		InputKey:    inputKey,
		Outputs:     getVideoOutputAssets(note),
	}

	return param
}
