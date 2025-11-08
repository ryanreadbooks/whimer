package concurrent

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	otelattribute "go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func SafeGo(job func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logx.ErrorStackf("panic: %v", err)
			}
		}()

		job()
	}()
}

const (
	tracerName             = "concurrent-async"
	unknownJobName         = "concurrent.job.unknown"
	traceSafeGoSpanName    = "concurrent.safego"
	traceDoneInCtxSpanName = "concurrent.donein"
	traceAttrJobName       = "job.name"
)

type SafeGo2Opt struct {
	Name       string
	Job        func(ctx context.Context) error
	LogOnError bool
}

func spanCtxWithoutCancel(parent context.Context, spanName, jobName string) (context.Context, oteltrace.Span) {
	parent = context.WithoutCancel(parent)
	return spanCtxFrom(parent, spanName, jobName)
}

func spanCtxFrom(parent context.Context, spanName, jobName string) (context.Context, oteltrace.Span) {
	spanCtx := trace.SpanContextFromContext(parent)
	newCtx := trace.ContextWithSpanContext(parent, spanCtx)
	tracer := otel.GetTracerProvider().Tracer(tracerName)

	newCtx, span := tracer.Start(
		newCtx,
		spanName,
		oteltrace.WithSpanKind(oteltrace.SpanKindInternal),
		oteltrace.WithAttributes(otelattribute.String(traceAttrJobName, jobName)),
	)

	return newCtx, span
}

func setSpanStatus(span oteltrace.Span, err error) {
	if err != nil {
		err = xerror.StripFrames(err)
		span.SetStatus(otelcodes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(otelcodes.Ok, otelcodes.Ok.String())
	}
}

func SafeGo2(ctx context.Context, opt SafeGo2Opt) {
	if opt.Job == nil {
		return
	}
	if opt.Name == "" {
		opt.Name = unknownJobName
	}

	go func() {
		newCtx, span := spanCtxWithoutCancel(ctx, traceSafeGoSpanName, opt.Name)
		defer func() {
			if err := recover(); err != nil {
				logx.ErrorStackf("panic: %v", err)
			}
			span.End()
		}()

		err := opt.Job(newCtx)
		if err != nil && opt.LogOnError {
			xlog.Msgf("concurrent job %s error", opt.Name).Err(err).Errorx(ctx)
		}

		setSpanStatus(span, err)
	}()
}

type DoneInJob func(ctx context.Context)

type DoneInJobWithError func(ctx context.Context) error

func DoneIn(duration time.Duration, job DoneInJob) {
	DoneInCtx(context.Background(), duration, job)
}

func DoneInCtx(parent context.Context, duration time.Duration, job DoneInJob) {
	SafeGo(func() {
		parent = context.WithoutCancel(parent)
		ctx, cancel := context.WithTimeout(parent, duration)
		defer cancel()

		job(ctx)
	})
}

type DoneInCtx2Opt struct {
	Name string
	Job  DoneInJobWithError
}

func DoneInCtx2(parent context.Context, duration time.Duration, opt DoneInCtx2Opt) {
	if opt.Job == nil {
		return
	}
	if opt.Name == "" {
		opt.Name = unknownJobName
	}

	SafeGo(func() {
		parent = context.WithoutCancel(parent)
		newCtx, cancel := context.WithTimeout(parent, duration)
		defer cancel()

		newCtx, span := spanCtxFrom(newCtx, traceDoneInCtxSpanName, opt.Name)
		defer span.End()

		err := opt.Job(newCtx)
		setSpanStatus(span, err)
	})
}
