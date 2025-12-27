package kafka

import (
	"net"
	"strings"
	"time"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/note/internal/config"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var (
	asyncWriter *xkafka.Writer // 异步写
	writer      *xkafka.Writer // 同步写

	publisher *Publisher
)

func Init(c *config.Config) {
	addrs := strings.Split(c.Kafka.Brokers, ",")
	transport := kafka.Transport{
		SASL: plain.Mechanism{
			Username: c.Kafka.Username,
			Password: c.Kafka.Password,
		},
		Dial: (&net.Dialer{
			Timeout:   3 * time.Second,
			DualStack: true,
		}).DialContext,
	}

	asyncWriter = xkafka.NewWriter(&kafka.Writer{
		Addr:      kafka.TCP(addrs...),
		Balancer:  &kafka.Hash{},
		Transport: &transport,
		Async:     true,
	})

	writer = xkafka.NewWriter(&kafka.Writer{
		Addr:      kafka.TCP(addrs...),
		Balancer:  &kafka.Hash{},
		Transport: &transport,
	})

	publisher = New(writer, asyncWriter)
}

func Close() {
	if asyncWriter != nil {
		asyncWriter.Close()
	}
	if writer != nil {
		writer.Close()
	}
}

func GetPublisher() *Publisher {
	return publisher
}
