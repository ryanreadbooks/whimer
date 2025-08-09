package srv

import (
	"context"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/infra"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index"
)

type DocumentService struct {
}

func (s *DocumentService) AddNoteTagDocs(ctx context.Context, tags []*searchv1.NoteTag) error {
	tgs := make([]*index.NoteTag, 0, len(tags))
	for _, t := range tags {
		tgs = append(tgs, &index.NoteTag{
			Id:    t.Id,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	var wg sync.WaitGroup
	err := xslice.BatchAsyncExec(&wg, tags, 200, func(start, end int) error {
		err := infra.EsDao().NoteTagIndexer.BulkAdd(ctx, tgs[start:end])
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "document note tag add failed").WithCtx(ctx)
	}

	return nil
}
