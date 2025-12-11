package handler

import (
	"fmt"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// CallbackPayload 回调请求体
type CallbackPayload struct {
	TaskId      string `json:"task_id"`
	Namespace   string `json:"namespace"`
	TaskType    string `json:"task_type"`
	State       string `json:"state"`
	OutputArgs  []byte `json:"output_args,omitempty,optional"`
	ErrorMsg    string `json:"error_msg,omitempty,optional"`
	TraceId     string `json:"trace_id,omitempty,optional"`
	CompletedAt int64  `json:"completed_at,omitempty,optional"`
}

// NoteProcessCallback 笔记处理流程回调
func (h *Handler) NoteProcessCallback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[CallbackPayload](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xlog.Msgf("note process callback received").
			Extras("req", fmt.Sprintf("%+v", req)).
			Extras("taskId", req.TaskId).
			Infox(r.Context())

		// TODO 处理回调
		err = h.Svc.NoteProcessSrv.Process(r.Context(), req.TaskId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
