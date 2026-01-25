package cache

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/note"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/recentcontact"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	noteStatStore      *notecache.StatStore
	recentContactStore *recentcontact.Store
)

func Init(c *config.Config, rd *redis.Redis) {
	recentcontact.Init(rd)
	noteStatStore = notecache.NewStatStore(rd, c.JobConfig.NoteEventJob.NumOfList)
	recentContactStore = recentcontact.New(rd)
}

func NoteStatStore() *notecache.StatStore {
	return noteStatStore
}

func RecentContactStore() *recentcontact.Store {
	return recentContactStore
}
