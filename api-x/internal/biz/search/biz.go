package search

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra/cache"
)

type SearchBiz struct {
	NoteStatSyncer *NoteStatSyncer
}

func NewSearchBiz(c *config.Config) *SearchBiz {
	b := &SearchBiz{
		NoteStatSyncer: &NoteStatSyncer{
			NoteCache: cache.NewNoteCache(c.JobConfig.NoteEventJob.NumOfList),
		},
	}

	return b
}
