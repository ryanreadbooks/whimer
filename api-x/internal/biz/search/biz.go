package search

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra/cache"
)

type Biz struct {
	NoteStatSyncer *NoteStatSyncer
}

func NewSearchBiz(c *config.Config) *Biz {
	b := &Biz{
		NoteStatSyncer: &NoteStatSyncer{
			NoteCache: cache.NewNoteCache(c.JobConfig.NoteEventJob.NumOfList),
		},
	}

	return b
}
