package transport

import (
	"net/http"

	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

func SpanTracing(next http.RoundTripper) http.RoundTripper {
	return Transporter(func(req *http.Request) (*http.Response, error) {
		ctx := req.Context()
		tracer := trace.TracerFromContext(ctx)
		propagater := otel.GetTextMapPropagator()

		spanName := req.URL.Path
		ctx, span := tracer.Start(ctx,
			spanName,
			oteltrace.WithSpanKind(oteltrace.SpanKindClient),
			oteltrace.WithAttributes(semconv.HTTPClientAttributesFromHTTPRequest(req)...))
		defer span.End()

		req = req.WithContext(ctx)
		propagater.Inject(ctx, propagation.HeaderCarrier(req.Header))

		resp, err := next.RoundTrip(req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return resp, err
		}

		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCodeAndSpanKind(resp.StatusCode, oteltrace.SpanKindClient))

		return resp, err
	})
}
