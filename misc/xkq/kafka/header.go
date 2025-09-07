package kafka

import (
	"context"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xkq/header"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

type HeaderCarrier map[string]string

// Implements go.opentelemetry.io/otel/propagation.TextMapCarrier
func (hc HeaderCarrier) Get(key string) string {
	v, _ := hc[key]
	return v
}

func (hc HeaderCarrier) Set(key string, value string) {
	hc[key] = value
}

func (hc HeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(hc))
	for k := range hc {
		keys = append(keys, k)
	}

	return keys
}

func (hc HeaderCarrier) KafkaHeader() []kafka.Header {
	hds := make([]kafka.Header, 0, len(hc))
	for k, v := range hc {
		hds = append(hds, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	return hds
}

func HeaderCarrierFromKafkaHeaders(hds []kafka.Header) HeaderCarrier {
	m := make(map[string]string, len(hds))
	for _, r := range hds {
		m[r.Key] = string(r.Value)
	}

	return m
}

func ContextFromKafkaHeaders(hds []kafka.Header) context.Context {
	carrier := HeaderCarrierFromKafkaHeaders(hds)
	ctx := context.Background()
	ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
	uid, _ := strconv.ParseInt(carrier.Get(header.HeaderUid), 10, 64)

	ctx = metadata.WithUid(ctx, uid)
	return ctx
}
