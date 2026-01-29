package messaging

import (
	"context"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	sysmsgkfkdao "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dao/kafka/sysmsg"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var (
	rootCtx    context.Context
	rootCancel context.CancelFunc
)

var (
	sysMsgDeletionConsumer *kafka.Reader
	noteEventConsumer      *kafka.Reader
)

func Init(c *config.Config, manager *app.Manager) {
	rootCtx, rootCancel = context.WithCancel(context.Background())
	addrs := strings.Split(c.Kafka.Brokers, ",")
	sysMsgDeletionConsumer = newKafkaReader(c, addrs, sysmsgkfkdao.DeletionTopic, sysmsgkfkdao.DeletionTopicGroup)
	noteEventConsumer = newKafkaReader(c, addrs, NoteEventTopic, NoteEventTopicGroupName)

	start(manager)
}

func start(manager *app.Manager) {
	startSysMsgDeletionConsumer(manager)
	startNoteEventConsumer(manager)
}

func Close() {
	rootCancel()
	if sysMsgDeletionConsumer != nil {
		sysMsgDeletionConsumer.Close()
	}
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
			CommitInterval:        time.Second * 1,
		},
	)

	return r
}
