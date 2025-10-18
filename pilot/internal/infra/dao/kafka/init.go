package kafka

import (
	"net"
	"strings"
	"time"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var (
	kafkaDao         *impl
	asyncKafkaWriter *xkafka.Writer
	syncKafkaWriter  *xkafka.Writer
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

	asyncKafkaWriter = xkafka.NewWriter(&kafka.Writer{
		Addr:      kafka.TCP(addrs...),
		Balancer:  &kafka.Hash{},
		Transport: &transport,
		Async:     true,
	})

	syncKafkaWriter = xkafka.NewWriter(&kafka.Writer{
		Addr:      kafka.TCP(addrs...),
		Balancer:  &kafka.Hash{},
		Transport: &transport,
	})

	kafkaDao = New(asyncKafkaWriter, syncKafkaWriter)
}

func Close() {
	if asyncKafkaWriter != nil {
		asyncKafkaWriter.Close()
	}
	if syncKafkaWriter != nil {
		syncKafkaWriter.Close()
	}
}
