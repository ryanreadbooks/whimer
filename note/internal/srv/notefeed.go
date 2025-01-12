package srv

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type NoteFeedSrv struct {
	Ctx *Service

	noteBiz         biz.NoteBiz
	noteCreatorBiz  biz.NoteCreatorBiz
	noteInteractBiz biz.NoteInteractBiz
}

func NewNoteFeedSrv(ctx *Service, biz biz.Biz) *NoteFeedSrv {
	s := &NoteFeedSrv{
		Ctx:             ctx,
		noteBiz:         biz.Note,
		noteCreatorBiz:  biz.Creator,
		noteInteractBiz: biz.Interact,
	}

	return s
}

// 信息流随机获取最多count条笔记
// TODO (refactor me in biz)
func (s *NoteFeedSrv) FeedRandomGet(ctx context.Context, count int32) (*model.Notes, error) {
	return s.randomGet(ctx, int(count))
}

func (s *NoteFeedSrv) randomGet(ctx context.Context, count int) (*model.Notes, error) {
	var (
		err    error
		lastId uint64
		wg     sync.WaitGroup
		items  []*dao.Note // items为随机获取的结果
	)

	wg.Add(1)
	concurrent.DoneIn(time.Second*10, func(sCtx context.Context) {
		defer wg.Done()
		//  TODO optimize by using local cache
		id, sErr := infra.Dao().NoteDao.GetPublicLastId(sCtx)
		if sErr != nil {
			xlog.Msg("note repo get public last id failed").Err(err).Errorx(sCtx)
		}
		lastId = id
	})

	// TODO optimize by using local cache
	maxCnt, err := infra.Dao().NoteDao.GetPublicCount(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "note repo get public count failed").WithCtx(ctx)
	}

	wg.Wait()

	itemsMap := make(map[uint64]*dao.Note, count)
	if maxCnt <= uint64(count) {
		// we fetch all
		items, err = infra.Dao().NoteDao.GetPublicAll(ctx)
		if err != nil {
			return nil, xerror.Wrapf(err, "note repo get public all failed").WithCtx(ctx).WithExtra("count", count)
		}
	} else {
		var notes []*dao.Note
		for tryCnt := 1; tryCnt <= 8 && len(itemsMap) < count; tryCnt++ {
			begin := rand.Int63n(int64(lastId))
			if begin == 0 {
				begin = 1
			}
			notes, err = infra.Dao().NoteDao.GetPublicByCursor(ctx, uint64(begin), count)
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

	result, err := s.noteBiz.AssembleNotes(ctx, model.NoteSliceFromDao(items))
	if err != nil {
		return nil, xerror.Wrapf(err, "feed srv assemble notes failed")
	}

	result, _ = s.noteInteractBiz.AssignNoteLikes(ctx, result)
	result, _ = s.noteInteractBiz.AssignNoteReplies(ctx, result)

	return result, nil
}

func (s *NoteFeedSrv) GetNoteDetail(ctx context.Context, noteId uint64) (*model.Note, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get note detail failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy == global.PrivacyPrivate && note.Owner != uid {
		return nil, global.ErrNoteNotPublic
	}

	res, err := s.noteBiz.AssembleNotes(ctx, note.AsSlice())
	if err != nil || len(res.Items) == 0 {
		return nil, xerror.Wrapf(err, "assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	res, _ = s.noteInteractBiz.AssignNoteLikes(ctx, res)
	res, _ = s.noteInteractBiz.AssignNoteReplies(ctx, res)
	return res.Items[0], nil
}
