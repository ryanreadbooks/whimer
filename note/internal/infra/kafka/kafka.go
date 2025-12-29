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

	pub *Publisher
)

func Writer() *xkafka.Writer {
	return writer
}

func AsyncWriter() *xkafka.Writer {
	return asyncWriter
}

type Publisher struct {
	w  *xkafka.Writer
	aw *xkafka.Writer
}

func GetPublisher() *Publisher {
	return pub
}

func (p *Publisher) Writer() *xkafka.Writer {
	return p.w
}

func (p *Publisher) AsyncWriter() *xkafka.Writer {
	return p.aw
}

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

	pub = &Publisher{
		w:  writer,
		aw: asyncWriter,
	}
}

func Close() {
	if asyncWriter != nil {
		asyncWriter.Close()
	}
	if writer != nil {
		writer.Close()
	}
}
