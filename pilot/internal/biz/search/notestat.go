package search

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/cache"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"

	// "github.com/ryanreadbooks/whimer/misc/xlog"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

type NoteStatSyncer struct {
	NoteCache *cache.NoteCache
}

func (s *NoteStatSyncer) AddLikeCount(ctx context.Context, noteId string, increment int64) error {
	return s.addStatCount(ctx, cache.NoteLikeCountStat, noteId, increment)
}

func (s *NoteStatSyncer) AddCommentCount(ctx context.Context, noteId string, increment int64) error {
	return s.addStatCount(ctx, cache.NoteCommentCountStat, noteId, increment)
}

func (s *NoteStatSyncer) addStatCount(ctx context.Context, statType cache.NoteInteractStatType,
	noteId string, increment int64) error {
	err := s.NoteCache.Add(ctx, statType, cache.NoteStatRepr{
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

// consume note like count event
func (s *NoteStatSyncer) PollLikeCount(ctx context.Context) error {
	stats, err := s.NoteCache.ConsumeLikeCount(ctx, 1)
	if err != nil {
		return xerror.Wrapf(err, "note stat syncer failed to poll like count").WithCtx(ctx)
	}

	// xlog.Msgf("note stat poll like count len = %d", len(stats)).Debugx(ctx)

	// remove duplicates
	reqs := s.removeDupAndDoMap(stats)
	if len(reqs) != 0 {
		_, err := dep.DocumentServer().BatchUpdateNoteLikeCount(ctx,
			&searchv1.BatchUpdateNoteLikeCountRequest{Counts: reqs})
		if err != nil {
			return xerror.Wrapf(err, "note stat syncer update note like count failed").WithCtx(ctx)
		}
	} else {
		// xlog.Msg("note stat poll like count result empty").Debugx(ctx)
	}

	return nil
}

// consume note comment count event
func (s *NoteStatSyncer) PollCommentCount(ctx context.Context) error {
	stats, err := s.NoteCache.ConsumeCommentCount(ctx, 1)
	if err != nil {
		return xerror.Wrapf(err, "note stat syncer failed to poll comment count").WithCtx(ctx)
	}

	// xlog.Msgf("note stat poll comment count len = %d", len(stats)).Debugx(ctx)

	reqs := s.removeDupAndDoMap(stats)
	if len(reqs) != 0 {
		_, err := dep.DocumentServer().BatchUpdateNoteCommentCount(ctx,
			&searchv1.BatchUpdateNoteCommentCountRequest{Counts: reqs})
		if err != nil {
			return xerror.Wrapf(err, "note stat syncer update note comment count failed").WithCtx(ctx)
		}
	} else {
		// xlog.Msg("note stat poll comment count result empty").Debugx(ctx)
	}

	return nil
}

func (s *NoteStatSyncer) removeDupAndDoMap(stats []cache.NoteStatRepr) map[string]int64 {
	tmp := make(map[string]int64, len(stats))
	for _, stat := range stats {
		tmp[stat.NoteId] += stat.Inc
	}

	res := make([]cache.NoteStatRepr, 0, len(stats))
	for noteId, incr := range tmp {
		if incr != 0 { // 0 means updatign to es is unnecessary
			res = append(res, cache.NoteStatRepr{
				NoteId: noteId,
				Inc:    incr,
			})
		}
	}

	reqs := make(map[string]int64, len(res))
	for _, s := range res {
		reqs[s.NoteId] = s.Inc
	}

	return reqs
}
