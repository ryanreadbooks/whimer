package svc

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"
	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"
)

type NoteFeedSvc struct {
	Ctx *ServiceContext
}

func NewNoteFeedSvc(ctx *ServiceContext) *NoteFeedSvc {
	s := &NoteFeedSvc{
		Ctx: ctx,
	}

	return s
}

// 信息流随机获取最多count条笔记
func (s *NoteFeedSvc) FeedRandomGet(ctx context.Context, count int32) (*notemodel.BatchNoteItem, error) {
	return s.randomGet(ctx, int(count))
}

func (s *NoteFeedSvc) randomGet(ctx context.Context, count int) (*notemodel.BatchNoteItem, error) {
	var (
		err    error
		lastId uint64
		wg     sync.WaitGroup
		items  []*noterepo.Model // items为随机获取的结果
	)

	wg.Add(1)
	concurrent.DoneIn(time.Second*20, func(sCtx context.Context) {
		defer wg.Done()
		//  TODO optimize by using local cache
		id, sErr := infra.Repo().NoteRepo.GetPublicLastId(sCtx)
		if sErr != nil {
			xlog.Msg("note repo get public last id failed").Err(err).Errorx(sCtx)
		}
		lastId = id
	})

	// TODO optimize by using local cache
	maxCnt, err := infra.Repo().NoteRepo.GetPublicCount(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "note repo get public count failed").WithCtx(ctx)
	}

	wg.Wait()

	itemsMap := make(map[uint64]*noterepo.Model, count)
	if maxCnt <= uint64(count) {
		// we fetch all
		items, err = infra.Repo().NoteRepo.GetPublicAll(ctx)
		if err != nil {
			return nil, xerror.Wrapf(err, "note repo get public all failed").WithCtx(ctx).WithExtra("count", count)
		}
	} else {
		var notes []*noterepo.Model
		for tryCnt := 1; tryCnt <= 8 && len(itemsMap) < count; tryCnt++ {
			begin := rand.Int63n(int64(lastId))
			if begin == 0 {
				begin = 1
			}
			notes, err = infra.Repo().NoteRepo.GetPublicByCursor(ctx, uint64(begin), count)
			if err != nil {
				return nil, xerror.Wrapf(err, "note repo get public by cursor failed").
					WithExtra("begin", begin).
					WithExtra("count", count).
					WithCtx(ctx)
			}
			for _, note := range notes {
				itemsMap[note.Id] = note
			}
		}
		items = maps.Values(itemsMap)
	}

	return AssembleNotes(ctx, items)
}
