package dao

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	kafkadao "github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka"
)

func Init(c *config.Config) {
	kafkadao.Init(c)
}

func Close() {
	kafkadao.Close()
}
