package handler

import (
	"encoding/json"
	"net/http"

	conductorsdk "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/task"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// NoteAssetCallbackPayload 回调请求体
type NoteAssetCallbackPayload struct {
	// in http query
	NoteId int64 `form:"note_id"`

	// in http req body
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
	// bucket + / + key 组成实际存储的asset_key
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	// 输出视频的实际参数
	Info *model.VideoInfo `json:"info"`
}

// NoteAssetProcessCallback 笔记处理流程回调
func (h *Handler) NoteAssetProcessCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteAssetCallbackPayload](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xlog.Msgf("note process callback received").
			Extras("taskId", req.TaskId).
			Infox(r.Context())

		var (
			ctx   = r.Context()
			input = &srv.HandleAssetProcessResultReq{
				NoteId:      req.NoteId,
				TaskId:      req.TaskId,
				Success:     req.State == conductorsdk.TaskStateSuccess,
				ErrorOutput: req.OutputArgs,
			}
		)

		if input.Success {
			var output NoteVideoAssetCallbackOutput
			// 成功时解析参数
			err = json.Unmarshal(req.OutputArgs, &output)
			if err != nil {
				// 打错误日志不错误退出
				xlog.Msg("note video asset process callback output parse failed").
					Err(err).
					Extras("taskId", req.TaskId).
					Extras("output_args", string(req.OutputArgs)).
					Errorx(ctx)
			} else {
				for _, item := range output.Outputs {
					input.VideoMetas = append(input.VideoMetas, &model.VideoAssetMetadata{
						Key:  item.Bucket + "/" + item.Key,
						Info: item.Info,
					})
				}
			}
		}

		// 处理回调
		err = h.Svc.NoteProcedureSrv.HandleCallbackAssetProcedureResult(ctx, input)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
