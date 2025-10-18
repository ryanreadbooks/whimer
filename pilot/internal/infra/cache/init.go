package cache

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/recentcontact"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func Init(c *config.Config, rd *redis.Redis) {
	recentcontact.Init(rd)
}
