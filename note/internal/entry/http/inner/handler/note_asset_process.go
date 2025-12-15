package handler

import (
	"encoding/json"
	"net/http"

	conductorsdk "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/task"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// NoteAssetCallbackPayload 回调请求体
type NoteAssetCallbackPayload struct {
	// in req query
	NoteId int64 `form:"note_id"`

	// in req body
	TaskId      string `json:"task_id"`
	Namespace   string `json:"namespace"`
	TaskType    string `json:"task_type"`
	State       string `json:"state"`
	OutputArgs  []byte `json:"output_args,omitempty,optional"`
	ErrorMsg    string `json:"error_msg,omitempty,optional"`
	TraceId     string `json:"trace_id,omitempty,optional"`
	CompletedAt int64  `json:"completed_at,omitempty,optional"`
}

type NoteVideoAssetCallbackOutput struct {
	Outputs []VideoOutput `json:"outputs"`
}

type VideoOutput struct {
	// Bucket 存储桶名称
	Bucket string `json:"bucket"`
	// Key 文件路径/键名
	Key string `json:"key"`
	// Info 输出视频的实际参数（通过 ffprobe 获取）
	Info *VideoInfo `json:"info"`
}

type VideoInfo struct {
	// Width 视频宽度（像素）
	Width int `json:"width"`
	// Height 视频高度（像素）
	Height int `json:"height"`
	// Duration 视频时长（秒）
	Duration float64 `json:"duration"`
	// Bitrate 总码率（bps）
	Bitrate int64 `json:"bitrate"`
	// Codec 视频编码器
	Codec string `json:"codec"`
	// Framerate 帧率
	Framerate float64 `json:"framerate"`
	// AudioCodec 音频编码器
	AudioCodec string `json:"audio_codec"`
	// AudioSampleRate 音频采样率（Hz）
	AudioSampleRate int `json:"audio_sample_rate"`
	// AudioChannels 音频声道数
	AudioChannels int `json:"audio_channels"`
	// AudioBitrate 音频码率（bps）
	AudioBitrate int64 `json:"audio_bitrate"`
}

// NoteAssetProcessCallback 笔记处理流程回调
func (h *Handler) NoteAssetProcessCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteAssetCallbackPayload](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		msg, _ := json.Marshal(req)
		var output NoteVideoAssetCallbackOutput
		err2 := json.Unmarshal(req.OutputArgs, &output)
		if err2 == nil {
			xlog.Msgf("video output args: %+v", output)
		}

		xlog.Msgf("note process callback received").
			Extras("req", string(msg)).
			Extras("taskId", req.TaskId).
			Infox(r.Context())

		var (
			ctx   = r.Context()
			input = &srv.HandleAssetProcessResultReq{
				NoteId:  req.NoteId,
				TaskId:  req.TaskId,
				Success: req.State == conductorsdk.TaskStateSuccess,
			}
		)

		// 处理回调
		err = h.Svc.NoteProcedureSrv.HandleAssetProcedureResult(ctx, input)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
