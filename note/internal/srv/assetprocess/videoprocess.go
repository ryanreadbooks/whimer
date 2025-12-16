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
type videoProcessReqParam struct {
	InputMode   string              `json:"input_mode"`
	InputBucket string              `json:"input_bucket"`
	InputKey    string              `json:"input_key"`
	Outputs     []*videoProcessDest `json:"outputs"`
}

type videoProcessDest struct {
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

func getVideoOutputAssets(note *model.Note) []*videoProcessDest {
	// key中包含了bucket 这里要将bucket去掉

	assets := make([]*videoProcessDest, 3)
	assets[0] = &videoProcessDest{
		Bucket:    note.Videos.H264.GetBucket(),
		OutputKey: note.Videos.H264.TrimBucket(),
		Settings: &EncodeSettings{
			VideoCodec: "libx264",
			Preset:     "medium",
			CRF:        26,
			MaxHeight:  1080,
		},
	}
	assets[1] = &videoProcessDest{
		Bucket:    note.Videos.H265.GetBucket(),
		OutputKey: note.Videos.H265.TrimBucket(),
		Settings: &EncodeSettings{
			VideoCodec: "libx265",
			Preset:     "medium",
			CRF:        28,
			MaxHeight:  1080,
		},
	}
	assets[2] = &videoProcessDest{
		Bucket:    note.Videos.AV1.GetBucket(),
		OutputKey: note.Videos.AV1.TrimBucket(),
		Settings: &EncodeSettings{
			VideoCodec: "libsvtav1",
			Preset:     "8",
			CRF:        35,
			MaxHeight:  1080,
		},
	}
	return assets
}

func (p *VideoProcessor) Process(ctx context.Context, note *model.Note) (string, error) {
	rawFileId := note.Videos.GetRawUrl()
	rawBucket := note.Videos.GetRawBucket()
	inputKey := strings.TrimPrefix(rawFileId, rawBucket+"/")

	param := &videoProcessReqParam{
		InputMode:   "s3",
		InputBucket: rawBucket,
		InputKey:    inputKey,
		Outputs:     getVideoOutputAssets(note),
	}
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
