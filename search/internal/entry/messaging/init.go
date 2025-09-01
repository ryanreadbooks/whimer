package messaging

import (
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/kafkadao"
	"github.com/ryanreadbooks/whimer/search/internal/srv"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var (
	noteEventReader      *kafka.Reader
	noteEventBatchReader *xkafka.BatchReader
)

func Init(c *config.Config, svc *srv.Service) {
	addrs := strings.Split(c.Kafka.Brokers, ",")
	noteEventReader = newKafkaReader(c, addrs, kafkadao.EsNoteTopic, kafkadao.EsNoteTopicGroup)
	noteEventBatchReader = xkafka.NewBatchReader(noteEventReader, xkafka.BatchReaderConfig{
		BatchSize:    c.ConsumerTopicConfig.Get(kafkadao.EsNoteTopic).BatchSize,
		BatchTimeout: c.ConsumerTopicConfig.Get(kafkadao.EsNoteTopic).BatchTimeout,
	})

	start(svc)
}

func start(svc *srv.Service) {
	startHandlingNoteEvents(svc)
}

func newKafkaReader(c *config.Config, addrs []string, topic, groupId string) *kafka.Reader {
	r := kafka.NewReader(
		kafka.ReaderConfig{
			Brokers: addrs,
			Topic:   topic,
			GroupID: groupId,
			Dialer: &kafka.Dialer{
				Timeout:   time.Second * 15,
				DualStack: true,
				SASLMechanism: plain.Mechanism{
					Username: c.Kafka.Username,
					Password: c.Kafka.Password,
				},
			},
			WatchPartitionChanges: true,
			CommitInterval:        time.Second * 1, // 异步提交offset
		},
	)

	return r
}

func Close() {
	noteEventReader.Close()
}
