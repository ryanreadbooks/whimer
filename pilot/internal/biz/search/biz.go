package search

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"

	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/note"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
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

func (b *Biz) AddNoteDoc(ctx context.Context, note *searchv1.Note) error {
	_, err := dep.DocumentServer().
		BatchAddNote(ctx,
			&searchv1.BatchAddNoteRequest{
				Notes: []*searchv1.Note{note},
			})
	if err != nil {
		return xerror.Wrapf(err, "add note doc failed").WithExtra("note_id", note.NoteId).WithCtx(ctx)
	}

	return nil
}

func (b *Biz) DeleteNoteDoc(ctx context.Context, noteId string) error {
	_, err := dep.DocumentServer().BatchDeleteNote(ctx, &searchv1.BatchDeleteNoteRequest{
		Ids: []string{noteId},
	})
	if err != nil {
		return xerror.Wrapf(err, "delete note doc failed").WithExtra("note_id", noteId).WithCtx(ctx)
	}

	return nil
}
