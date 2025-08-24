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
	indexTags := make([]*index.NoteTag, 0, len(tags))
	for _, t := range tags {
		indexTags = append(indexTags, &index.NoteTag{
			Id:    t.Id,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	controllableExec(ctx, s.pool, indexTags, func(ctx context.Context, datas []*index.NoteTag) error {
		err := infra.EsDao().NoteTagIndexer.BulkAdd(ctx, datas)
		if err != nil {
			xlog.Msg("document note tag add failed").Err(err).Errorx(ctx)
			return err
		}

		return nil
	})

	return nil
}

func (s *DocumentService) AddNoteDocs(ctx context.Context, notes []*searchv1.Note) error {
	indexNotes := make([]*index.Note, 0, len(notes))
	for _, n := range notes {
		indexTags := make([]*index.NoteTag, 0, len(n.GetTagList()))
		for _, t := range n.GetTagList() {
			indexTags = append(indexTags, &index.NoteTag{
				Id:    t.Id,
				Name:  t.Name,
				Ctime: t.Ctime,
			})
		}

		indexNotes = append(indexNotes, &index.Note{
			NoteId:   n.NoteId,
			Title:    n.Title,
			Desc:     n.Desc,
			CreateAt: n.CreateAt,
			UpdateAt: n.UpdateAt,
			Author: index.NoteAuthor{
				Uid:      n.Author.Uid,
				Nickname: n.Author.Nickname,
			},
			TagList:    indexTags,
			AssetType:  index.NoteAssetConverter[n.GetAssetType()],
			Visibility: index.NoteVisibilityConverter[n.GetVisibility()],
		})
	}

	controllableExec(ctx, s.pool, indexNotes, func(ctx context.Context, datas []*index.Note) error {
		err := infra.EsDao().NoteIndexer.BulkAdd(ctx, datas)
		if err != nil {
			xlog.Msg("document note add failed").Err(err).Errorx(ctx)
			return err
		}

		return nil
	})

	return nil
}

func (s *DocumentService) DeleteNoteDocs(ctx context.Context, ids []string) error {
	controllableExec(ctx, s.pool, ids, func(ctx context.Context, datas []string) error {
		err := infra.EsDao().NoteIndexer.BulkDelete(ctx, datas)
		if err != nil {
			xlog.Msg("document note delete failed").Err(err).Errorx(ctx)
			return err
		}

		return nil
	})

	return nil
}

func controllableExec[T any](ctx context.Context,
	pool *ants.Pool,
	datas []T,
	job func(ctx context.Context, datas []T) error) {

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "controllable_exec",
		Job: func(newCtx context.Context) error {
			// 借助pool来控制写入速度
			errSubmit := pool.Submit(func() {
				errExec := xslice.BatchExec(datas, 200, func(start, end int) error {
					errJob := job(newCtx, datas[start:end])
					if errJob != nil {
						return errJob
					}

					return nil
				})

				if errExec != nil {
					xlog.Msg("batch exec failed").Err(errExec).Errorx(newCtx)
				}
			})

			if errSubmit != nil {
				xlog.Msg("pool submit failed").Err(errSubmit).Errorx(newCtx)
				return errSubmit
			}

			return nil

		},
	})
}
