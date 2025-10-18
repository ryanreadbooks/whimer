package search

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/note"
)

type Biz struct {
	NoteStatSyncer *NoteStatSyncer
}

func NewSearchBiz(c *config.Config) *Biz {
	b := &Biz{
		NoteStatSyncer: &NoteStatSyncer{
			NoteCache: notecache.New(infra.Cache(), c.JobConfig.NoteEventJob.NumOfList),
		},
	}

	return b
}
