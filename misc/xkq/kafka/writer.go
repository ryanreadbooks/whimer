package kafka

import (
	"context"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xkq/header"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

type Writer struct {
	w *kafka.Writer
}

func NewWriter(w *kafka.Writer) *Writer {
	return &Writer{
		w: w,
	}
}

const (
	tracerName            = "kafka"
	writeMessagesSpanName = "kafka.batch.write"
)

func (w *Writer) WriteMessage(ctx context.Context, msg kafka.Message) error {
	// 只往header中加入metadata
	var (
		uid        = metadata.Uid(ctx)
		tracer     = otel.GetTracerProvider().Tracer(tracerName)
		propagator = otel.GetTextMapPropagator()
	)

	// trace
	spanCtx, span := tracer.Start(
		ctx, writeMessagesSpanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(semconv.MessagingKafkaMessageKey(string(msg.Key))),
		trace.WithAttributes(semconv.MessagingSystemKafka),
	)
	defer span.End()

	headerMap := make(map[string]string)
	headerMap[header.HeaderUid] = strconv.FormatInt(uid, 10)
	headerCarrier := HeaderCarrier(headerMap)
	propagator.Inject(spanCtx, headerCarrier)

	msg.Headers = headerCarrier.KafkaHeader()
	err := w.w.WriteMessages(spanCtx, msg)
	setSpanStatus(span, err)
	return err
}

// 封装WriterMessages 主要是加入trace和一些metadata信息
func (w *Writer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	// 只往header中加入metadata
	var (
		uid        = metadata.Uid(ctx)
		tracer     = otel.GetTracerProvider().Tracer(tracerName)
		propagator = otel.GetTextMapPropagator()
	)

	// trace
	spanCtx, span := tracer.Start(
		ctx, writeMessagesSpanName,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.MessagingSystemKafka,
			attribute.KeyValue{
				Key:   semconv.MessagingDestinationPublishNameKey,
				Value: attribute.StringValue(msgs[0].Topic),
			},
		),
	)
	defer span.End()

	headerMap := make(map[string]string)
	headerMap[header.HeaderUid] = strconv.FormatInt(uid, 10)
	headerCarrier := HeaderCarrier(headerMap)
	propagator.Inject(spanCtx, headerCarrier)

	kafkaHeaders := headerCarrier.KafkaHeader()

	for idx := range len(msgs) {
		msgs[idx].Headers = kafkaHeaders
	}

	err := w.w.WriteMessages(spanCtx, msgs...)
	setSpanStatus(span, err)
	return err
}

func setSpanStatus(span trace.Span, err error) {
	if err != nil {
		err = xerror.StripFrames(err)
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	} else {
		span.SetStatus(codes.Ok, codes.Ok.String())
	}
}

func (w *Writer) Close() error {
	return w.w.Close()
}

func (w *Writer) Stats() kafka.WriterStats {
	return w.w.Stats()
}
