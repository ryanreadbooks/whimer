package handler

import (
	"fmt"
	"net/http"

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

// NoteAssetProcessCallback 笔记处理流程回调
func (h *Handler) NoteAssetProcessCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteAssetCallbackPayload](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xlog.Msgf("note process callback received").
			Extras("req", fmt.Sprintf("%+v", req)).
			Extras("taskId", req.TaskId).
			Infox(r.Context())

		// 处理回调
		err = h.Svc.NoteProcedureSrv.HandleAssetProcessResult(r.Context(), &srv.HandleAssetProcessResultReq{
			NoteId: req.NoteId,
			TaskId: req.TaskId,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
