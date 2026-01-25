package infra

import (
	"sync"

	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
	infracache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dao"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/repo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/syncjob"
)

var initOnce sync.Once

func Init(c *config.Config) {
	initOnce.Do(func() {
		initCache(c)
		dao.Init(c, Cache())
		dep.Init(c)
		initMisc(c)
		adapter.Init(c, Cache())
		repo.Init(infracache.RecentContactStore())
		syncjob.InitNoteStatSyncJob(syncjob.NoteStatSyncJobConfig{
			Tick:      c.JobConfig.NoteEventJob.Interval,
			NumOfList: int(c.JobConfig.NoteEventJob.NumOfList),
		}, dep.DocumentServer(), infracache.NoteStatStore())

		syncjob.GetNoteStatSyncJob().Start()
	})
}

func Close() {
	syncjob.GetNoteStatSyncJob().Stop()
	dao.Close()
}
