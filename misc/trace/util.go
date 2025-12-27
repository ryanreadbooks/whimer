package trace

import (
	"context"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

func SpanIDFromCtx(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasSpanID() {
		return spanCtx.SpanID().String()
	}

	return ""
}

func TraceIDFromCtx(ctx context.Context) string {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		return spanCtx.TraceID().String()
	}

	return ""
}

// ContextWithTraceparent 从 traceparent 字符串创建带 trace 信息的 context
// traceparent 格式: 00-{trace-id}-{span-id}-{flags}
func ContextWithTraceparent(ctx context.Context, traceparent string) context.Context {
	if traceparent == "" {
		return ctx
	}

	// 解析 traceparent: 00-{trace-id}-{span-id}-{flags}
	parts := strings.Split(traceparent, "-")
	if len(parts) != 4 {
		return ctx
	}

	traceIdHex := parts[1]
	spanIdHex := parts[2]

	traceId, err := trace.TraceIDFromHex(traceIdHex)
	if err != nil {
		return ctx
	}

	spanId, err := trace.SpanIDFromHex(spanIdHex)
	if err != nil {
		return ctx
	}

	// 创建 SpanContext
	spanCtx := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    traceId,
		SpanID:     spanId,
		TraceFlags: trace.FlagsSampled,
		Remote:     true,
	})

	return trace.ContextWithSpanContext(ctx, spanCtx)
}
