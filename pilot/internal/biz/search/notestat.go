package search

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/note"
)

type NoteInteractStatSyncer struct {
	NoteCache *notecache.StatStore
}

func (s *NoteInteractStatSyncer) AddCommentCount(ctx context.Context, noteId string, increment int64) error {
	return s.addStatCount(ctx, notecache.NoteCommentCountStat, noteId, increment)
}

func (s *NoteInteractStatSyncer) addStatCount(ctx context.Context, statType notecache.NoteInteractStatType,
	noteId string, increment int64,
) error {
	err := s.NoteCache.Add(ctx, statType, notecache.NoteStatRepr{
		NoteId: noteId,
		Inc:    increment,
	})
	if err != nil {
		return xerror.Wrapf(err, "failed to cache add note stat").
			WithExtras("note_id", noteId, "inc", increment).
			WithCtx(ctx)
	}

	return nil
}
