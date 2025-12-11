package srv

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/model/convert"
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
func (s *NoteFeedSrv) FeedRandomGet(ctx context.Context, count int32) (*model.Notes, error) {
	return s.randomGet(ctx, int(count))
}

// TODO optimize
func (s *NoteFeedSrv) randomGet(ctx context.Context, count int) (*model.Notes, error) {
	var (
		err    error
		lastId int64
		wg     sync.WaitGroup
		items  []*notedao.NotePO // items为随机获取的结果
	)

	wg.Add(1)
	concurrent.DoneInCtx(ctx, time.Second*10, func(sCtx context.Context) {
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

	itemsMap := make(map[int64]*notedao.NotePO, count)
	if maxCnt <= int64(count) {
		// we fetch all
		items, err = infra.Dao().NoteDao.GetPublicAll(ctx)
		if err != nil {
			return nil, xerror.Wrapf(err, "note repo get public all failed").WithCtx(ctx).WithExtra("count", count)
		}
	} else {
		var notes []*notedao.NotePO
		for tryCnt := 1; tryCnt <= 8 && len(itemsMap) < count; tryCnt++ {
			begin := rand.Int63n(int64(lastId))
			if begin == 0 {
				begin = 1
			}
			notes, err = infra.Dao().NoteDao.GetPublicByCursor(ctx, int64(begin), count)
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

	result, err := s.noteBiz.AssembleNotes(ctx, convert.NoteSliceFromDao(items))
	if err != nil {
		return nil, xerror.Wrapf(err, "feed srv assemble notes failed")
	}

	result, _ = s.noteInteractBiz.AssignNoteLikes(ctx, result)
	result, _ = s.noteInteractBiz.AssignNoteReplies(ctx, result)

	return result, nil
}

// 获取笔记详情 不包含private范围的笔记
func (s *NoteFeedSrv) GetNoteDetail(ctx context.Context, noteId int64) (*model.Note, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	note, err := s.noteBiz.GetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get note detail failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	if note.Privacy == model.PrivacyPrivate && note.Owner != uid {
		return nil, global.ErrNoteNotPublic
	}

	res := &model.Notes{Items: []*model.Note{note}}

	// 详细信息需要ext
	err = s.noteBiz.AssembleNotesExt(ctx, res.Items)
	if err != nil {
		return nil, xerror.Wrapf(err, "assemble notes ext failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	res, _ = s.noteInteractBiz.AssignNoteLikes(ctx, res)
	res, _ = s.noteInteractBiz.AssignNoteReplies(ctx, res)

	return res.Items[0], nil
}

func (s *NoteFeedSrv) GetNoteAuthor(ctx context.Context, noteId int64) (int64, error) {
	return s.noteBiz.GetNoteOwner(ctx, noteId)
}

// 批量获取笔记详情 不包含private范围的笔记
func (s *NoteFeedSrv) BatchGetNoteDetail(ctx context.Context, noteIds []int64) (map[int64]*model.Note, error) {
	notes, err := s.noteBiz.BatchGetNote(ctx, noteIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "batch get note failed").WithCtx(ctx)
	}

	if len(notes) == 0 {
		return map[int64]*model.Note{}, nil
	}

	// 过滤掉private的笔记
	pNotes := xmap.Filter(notes, func(k int64, v *model.Note) bool {
		return v.Privacy == model.PrivacyPrivate
	})

	nns := &model.Notes{Items: xmap.Values(pNotes)}
	err = s.noteBiz.AssembleNotesExt(ctx, nns.Items)
	if err != nil {
		return nil, xerror.Wrapf(err, "assemble notes ext failed").WithCtx(ctx)
	}

	nns, _ = s.noteInteractBiz.AssignNoteLikes(ctx, nns)
	nns, _ = s.noteInteractBiz.AssignNoteReplies(ctx, nns)

	resp := make(map[int64]*model.Note, len(nns.Items))
	for _, n := range nns.Items {
		resp[n.NoteId] = n
	}

	return resp, nil
}

// 获取用户最近发布的笔记
func (s *NoteFeedSrv) GetUserRecentNotes(ctx context.Context, user int64, maxCount int32) (*model.Notes, error) {
	result, err := s.noteBiz.GetUserRecentNote(ctx, user, maxCount)
	if err != nil {
		return nil, xerror.Wrapf(err, "feed srv failed to get user recent notes")
	}

	result, _ = s.noteInteractBiz.AssignNoteLikes(ctx, result)
	result, _ = s.noteInteractBiz.AssignNoteReplies(ctx, result)

	return result, nil
}

// 列出用户公开的笔记
func (s *NoteFeedSrv) ListUserPublicNotes(ctx context.Context, user int64, cursor int64, count int32) (*model.Notes,
	model.PageResult, error) {
	result, page, err := s.noteBiz.ListUserPublicNote(ctx, user, cursor, count)
	if err != nil {
		return nil, page, xerror.Wrapf(err, "feed srv failed to lsit user public note")
	}

	result, _ = s.noteInteractBiz.AssignNoteLikes(ctx, result)
	result, _ = s.noteInteractBiz.AssignNoteReplies(ctx, result)
	return result, page, nil
}

func (s *NoteFeedSrv) GetTagInfo(ctx context.Context, id int64) (*model.NoteTag, error) {
	tag, err := s.noteBiz.GetTag(ctx, id)
	if err != nil {
		return nil, xerror.Wrapf(err, "note biz failed to get tag")
	}

	return tag, nil
}

func (s *NoteFeedSrv) GetUserPublicPostedCount(ctx context.Context, user int64) (int64, error) {
	cnt, err := s.noteCreatorBiz.GetUserPublicPostedCount(ctx, user)
	if err != nil {
		return cnt, xerror.Wrapf(err, "feed srv get public posted cnt failed").WithCtx(ctx)
	}
	return cnt, nil
}

func (s *NoteFeedSrv) BatchCheckNoteExistence(ctx context.Context, noteIds []int64) (map[int64]bool, error) {
	var (
		reqUid = metadata.Uid(ctx)
	)

	notes, err := s.noteBiz.BatchGetNoteWithoutAsset(ctx, noteIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "note biz failed to batch get")
	}

	var result = make(map[int64]bool, len(notes))
	for _, noteId := range noteIds {
		if target, ok := notes[noteId]; ok {
			// 如果笔记存在，公开笔记所有人都可访问，私有笔记只有所有者可访问
			result[noteId] = target.Privacy != model.PrivacyPrivate || target.Owner == reqUid
		} else {
			result[noteId] = false
		}
	}

	return result, nil
}
