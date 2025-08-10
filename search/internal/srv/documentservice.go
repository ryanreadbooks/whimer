package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/infra"
	"github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index"

	"github.com/panjf2000/ants/v2"
)

type DocumentService struct {
	pool *ants.Pool
}

func NewDocumentService() *DocumentService {
	pool, _ := ants.NewPool(100)
	d := &DocumentService{
		pool: pool,
	}

	return d
}

func (s *DocumentService) Close() {
	s.pool.Release()
}

// 异步提交写入
func (s *DocumentService) AddNoteTagDocs(ctx context.Context, tags []*searchv1.NoteTag) error {
	tgs := make([]*index.NoteTag, 0, len(tags))
	for _, t := range tags {
		tgs = append(tgs, &index.NoteTag{
			Id:    t.Id,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	// 借助pool来控制写入速度
	concurrent.SafeGo2(ctx, func(newCtx context.Context) {
		errSubmit := s.pool.Submit(func() {
			errExec := xslice.BatchExec(tags, 200, func(start, end int) error {
				errAdd := infra.EsDao().NoteTagIndexer.BulkAdd(newCtx, tgs[start:end])
				if errAdd != nil {
					return errAdd
				}
				return nil
			})

			if errExec != nil {
				xlog.Msg("document note tag add failed").Err(errExec).Errorx(newCtx)
			}
		})

		if errSubmit != nil {
			xlog.Msg("document submit notetag add task failed").Err(errSubmit).Errorx(newCtx)
		}
	})

	return nil
}
