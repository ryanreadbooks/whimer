package kafka

import (
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka/sysmsg"
)

type impl struct {
	asyncK *xkafka.Writer
	syncK  *xkafka.Writer

	SysMsgEventProducer *sysmsg.SysMsgProducer
}

func New(asyncK, syncK *xkafka.Writer) *impl {
	return &impl{
		asyncK:              asyncK,
		syncK:               syncK,
		SysMsgEventProducer: sysmsg.NewProducer(asyncK, syncK),
	}
}

func Dao() *impl {
	return kafkaDao
}
