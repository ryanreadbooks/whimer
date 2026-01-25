package cache

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/note"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/recentcontact"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var noteStatStore *notecache.StatStore

func Init(c *config.Config, rd *redis.Redis) {
	recentcontact.Init(rd)
	noteStatStore = notecache.NewStatStore(rd, c.JobConfig.NoteEventJob.NumOfList)
}

func NoteStatStore() *notecache.StatStore {
	return noteStatStore
}
