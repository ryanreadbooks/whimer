package search

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"

	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/note"
)

type Biz struct {
	NoteStatSyncer *NoteInteractStatSyncer
}

func NewSearchBiz(c *config.Config) *Biz {
	b := &Biz{
		NoteStatSyncer: &NoteInteractStatSyncer{
			NoteCache: notecache.NewStatStore(infra.Cache(), c.JobConfig.NoteEventJob.NumOfList),
		},
	}
	return b
}
