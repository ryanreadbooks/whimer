package infra

import (
	"net"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao"
	"github.com/ryanreadbooks/whimer/search/internal/infra/kafkadao"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"

	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"
)

var (
	esDao            *esdao.EsDao
	kfkDao           *kafkadao.KafkaDao
	asyncKafkaWriter *xkafka.Writer
)

func Init(c *config.Config) {
	esDao = esdao.MustNew(c)
	esDao.MustInit(c)

	initKafka(c)
}

func EsDao() *esdao.EsDao {
	return esDao
}

func KafkaDao() *kafkadao.KafkaDao {
	return kfkDao
}

func Close() {
	asyncKafkaWriter.Close()
}

func initKafka(c *config.Config) {
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
		Async:     true, // 异步写入 允许写不入
	})

	kfkDao = kafkadao.New(asyncKafkaWriter)
}
