package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/trace"
	"github.com/ryanreadbooks/whimer/misc/xhttp/client"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// CallbackPayload 回调请求体
type CallbackPayload struct {
	TaskId      string          `json:"task_id"`
	Namespace   string          `json:"namespace"`
	TaskType    string          `json:"task_type"`
	State       model.TaskState `json:"state"`
	OutputArgs  []byte          `json:"output_args,omitempty"`
	ErrorMsg    string          `json:"error_msg,omitempty"`
	TraceId     string          `json:"trace_id,omitempty"`
	CompletedAt int64           `json:"completed_at"`
}

// CallbackBiz 回调业务逻辑
type CallbackBiz struct {
	httpcli *http.Client
}

// NewCallbackBiz 创建回调业务逻辑
func NewCallbackBiz() *CallbackBiz {
	// 使用 builder 构建带重试和链路追踪的 HTTP 客户端
	cli := client.NewBuilder().
		WithTimeout(10 * time.Second).
		WithDefaultRetry().
		WithTracing(). // 注入 trace id 到请求头
		Build()

	return &CallbackBiz{
		httpcli: cli,
	}
}

// TriggerCallback 触发回调（异步执行）
func (b *CallbackBiz) TriggerCallback(ctx context.Context, callbackUrl string, payload *CallbackPayload) {
	if callbackUrl == "" {
		return
	}

	// 从 payload.TraceId 创建 SpanContext 并注入到 context 中
	// traceid 格式为 W3C traceparent: 00-{trace-id}-{span-id}-{flags}
	traceCtx := trace.ContextWithTraceparent(ctx, payload.TraceId)

	// TODO 引入协程池 控制并发数量
	concurrent.SafeGo2(traceCtx, concurrent.SafeGo2Opt{
		Name:       "conductor.callback.trigger",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			b.doCallback(ctx, callbackUrl, payload)
			return nil
		},
	})
}

func (b *CallbackBiz) doCallback(ctx context.Context, callbackUrl string, payload *CallbackPayload) {
	body, err := json.Marshal(payload)
	if err != nil {
		xlog.Msg("callback marshal payload failed").
			Extras("callbackUrl", callbackUrl, "taskId", payload.TaskId).
			Err(err).Errorx(ctx)
		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, callbackUrl, bytes.NewReader(body))
	if err != nil {
		xlog.Msg("callback create request failed").
			Extras("callbackUrl", callbackUrl, "taskId", payload.TaskId).
			Err(err).Errorx(ctx)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := b.httpcli.Do(req)
	if err != nil {
		xlog.Msg("callback request failed").
			Extras("callbackUrl", callbackUrl, "taskId", payload.TaskId).
			Err(err).Errorx(ctx)
		return
	}
	defer resp.Body.Close()

	// 2xx 视为成功
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		xlog.Msg("callback success").
			Extras("callbackUrl", callbackUrl,
				"taskId", payload.TaskId,
				"statusCode", resp.StatusCode,
				"responseBody", resp.Body).
			Infox(ctx)
		return
	}

	xlog.Msg("callback failed with status").
		Extras("callbackUrl", callbackUrl,
			"taskId", payload.TaskId,
			"statusCode", resp.StatusCode).
		Errorx(ctx)
}
