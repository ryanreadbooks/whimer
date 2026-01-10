package data

import (
	"context"
	"maps"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// NoteData 笔记数据层 - 协调数据库和缓存操作
type NoteData struct {
	repo  *notedao.NoteRepo
	cache *notedao.NoteCache
}

func NewNoteData(repo *notedao.NoteRepo, cache *notedao.NoteCache) *NoteData {
	return &NoteData{
		repo:  repo,
		cache: cache,
	}
}

// FindOne 获取单个笔记，支持缓存选项
func (d *NoteData) FindOne(ctx context.Context, id int64, opts ...GetOptionFunc) (*notedao.NotePO, error) {
	opt := ApplyOptions(opts...)

	// 尝试从缓存获取
	if opt.UseCache {
		if cached, err := d.cache.GetNote(ctx, id); err == nil && cached != nil {
			return cached, nil
		}
	}

	// 从数据库获取
	note, err := d.repo.FindOne(ctx, id)
	if err != nil {
		return nil, err
	}

	// 异步回填缓存
	if opt.SetCache {
		concurrent.SafeGo(func() {
			if err2 := d.cache.SetNote(context.WithoutCancel(ctx), note); err2 != nil {
				xlog.Msg("note data failed to set cache when finding").
					Extras("noteId", note.Id).Err(err2).Errorx(ctx)
			}
		})
	}

	return note, nil
}

// FindOneForUpdate 获取单个笔记用于更新（带行锁，不使用缓存）
func (d *NoteData) FindOneForUpdate(ctx context.Context, id int64) (*notedao.NotePO, error) {
	return d.repo.FindOneForUpdate(ctx, id)
}

// BatchGet 批量获取笔记，支持缓存选项
func (d *NoteData) BatchGet(ctx context.Context, ids []int64, opts ...GetOptionFunc) (map[int64]*notedao.NotePO, error) {
	if len(ids) == 0 {
		return map[int64]*notedao.NotePO{}, nil
	}

	opt := ApplyOptions(opts...)
	result := make(map[int64]*notedao.NotePO, len(ids))

	// 尝试从缓存获取
	var missingIds []int64
	if opt.UseCache {
		cached, err := d.cache.MGetNotes(ctx, ids)
		if err != nil {
			xlog.Msg("note data failed to mget from cache").Err(err).Infox(ctx)
		}

		maps.Copy(result, cached)

		// 找出缺失的ID
		for _, id := range ids {
			if _, ok := result[id]; !ok {
				missingIds = append(missingIds, id)
			}
		}
	} else {
		missingIds = ids
	}

	// 从数据库获取缺失的
	if len(missingIds) > 0 {
		notes, err := d.repo.BatchGet(ctx, missingIds)
		if err != nil {
			return nil, err
		}

		for _, note := range notes {
			result[note.Id] = note
		}

		// 异步回填缓存
		if opt.SetCache && len(notes) > 0 {
			concurrent.SafeGo(func() {
				if err2 := d.cache.MSetNotes(context.WithoutCancel(ctx), notes); err2 != nil {
					xlog.Msg("note data failed to mset cache").Err(err2).Errorx(ctx)
				}
			})
		}
	}

	return result, nil
}

// GetNoteType 获取笔记类型
func (d *NoteData) GetNoteType(ctx context.Context, id int64) (model.NoteType, error) {
	return d.repo.GetNoteType(ctx, id)
}

// Insert 插入笔记
func (d *NoteData) Insert(ctx context.Context, note *notedao.NotePO) (int64, error) {
	id, err := d.repo.Insert(ctx, note)
	if err != nil {
		return 0, xerror.Wrapf(err, "note dao insert failed")
	}

	d.delNoteRelatedCache(ctx, id, note.Owner)

	return id, nil
}

func (d *NoteData) delNoteRelatedCache(ctx context.Context, noteId, ownerId int64) {
	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := d.cache.DelNoteRelatedCache(ctx, noteId, ownerId); err != nil {
			xlog.Msg("note data failed to del cache when updating").
				Extras("noteId", noteId).Err(err).Errorx(ctx)
		}
	})
}

// Update 更新笔记，同时删除相关缓存
func (d *NoteData) Update(ctx context.Context, note *notedao.NotePO) error {
	err := d.repo.Update(ctx, note)
	if err != nil {
		return err
	}

	// 异步删除缓存
	d.delNoteRelatedCache(ctx, note.Id, note.Owner)
	return nil
}

// Delete 删除笔记，同时删除相关缓存
func (d *NoteData) Delete(ctx context.Context, noteId int64) error {
	// 删除笔记
	err := d.repo.Delete(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "note dao delete failed")
	}

	d.delNoteCache(ctx, noteId)

	return nil
}

func (d *NoteData) DeleteNote(ctx context.Context, po *notedao.NotePO) error {
	err := d.repo.Delete(ctx, po.Id)
	if err != nil {
		return xerror.Wrapf(err, "note dao delete failed")
	}

	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		d.cache.DelNoteRelatedCache(ctx, po.Id, po.Owner)
	})

	return nil
}

func (d *NoteData) delNoteCache(ctx context.Context, noteId int64) {
	// 同步调用获取owner
	ownerId, err := d.repo.GetOwner(ctx, noteId)
	if err != nil {
		xlog.Msg("note data failed to get owner when deleting cache").
			Extras("noteId", noteId).Err(err).Errorx(ctx)
	}

	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		d.cache.DelNoteRelatedCache(ctx, noteId, ownerId)
	})
}

// UpdateState 更新笔记状态，同时删除缓存
func (d *NoteData) UpdateState(
	ctx context.Context,
	noteId int64,
	state model.NoteState,
) error {
	err := d.repo.UpdateState(ctx, noteId, state)
	if err != nil {
		return err
	}

	d.delNoteCache(ctx, noteId)

	return nil
}

// 带条件检查的笔记状态更新 同时删除缓存
//
// 如果当前状态小于等于目标状态，则更新状态，否则不更新
func (d *NoteData) UpgradeState(
	ctx context.Context,
	noteId int64,
	state model.NoteState,
) error {
	err := d.repo.UpgradeState(ctx, noteId, state)
	if err != nil {
		return err
	}

	d.delNoteCache(ctx, noteId)

	return nil
}

// GetPostedCountByOwner 获取用户发布的笔记数量，支持缓存选项
func (d *NoteData) GetPostedCountByOwner(ctx context.Context, uid int64, opts ...GetOptionFunc) (int64, error) {
	opt := ApplyOptions(opts...)

	// 尝试从缓存获取
	if opt.UseCache {
		if cnt, ok, err := d.cache.GetNoteCountByOwner(ctx, uid); err == nil && ok {
			return cnt, nil
		}
	}

	// 从数据库获取
	cnt, err := d.repo.Count(ctx, WithNoteOwnerEqual(uid), WithNoteStateEqual(model.NoteStatePublished))
	if err != nil {
		return 0, err
	}

	// 异步回填缓存
	if opt.SetCache {
		concurrent.SafeGo(func() {
			if err2 := d.cache.SetNoteCountByOwner(context.WithoutCancel(ctx), uid, cnt); err2 != nil {
				xlog.Msg("note data failed to set count cache").
					Extras("uid", uid).Err(err2).Errorx(ctx)
			}
		})
	}

	return cnt, nil
}

func (d *NoteData) GetCountWithStateByOwner(ctx context.Context, uid int64, states ...model.NoteState) (int64, error) {
	return d.repo.Count(ctx, WithNoteOwnerEqual(uid), WithNoteStateIn(states...))
}

// GetPublicPostedCountByOwner 获取用户公开发布的笔记数量，支持缓存选项
func (d *NoteData) GetPublicPostedCountByOwner(ctx context.Context, uid int64, opts ...GetOptionFunc) (int64, error) {
	opt := ApplyOptions(opts...)

	// 尝试从缓存获取
	if opt.UseCache {
		if cnt, ok, err := d.cache.GetPublicNoteCountByOwner(ctx, uid); err == nil && ok {
			return cnt, nil
		}
	}

	// 从数据库获取
	cnt, err := d.repo.Count(ctx,
		WithNoteOwnerEqual(uid),
		WithNotePrivacyEqual(model.PrivacyPublic),
		WithNoteStateEqual(model.NoteStatePublished),
	)
	if err != nil {
		return 0, err
	}

	// 异步回填缓存
	if opt.SetCache {
		concurrent.SafeGo(func() {
			if err2 := d.cache.SetPublicNoteCountByOwner(context.WithoutCancel(ctx), uid, cnt); err2 != nil {
				xlog.Msg("note data failed to set public count cache").
					Extras("uid", uid).Err(err2).Errorx(ctx)
			}
		})
	}

	return cnt, nil
}

// List 查询笔记列表（支持 conditions）
func (d *NoteData) List(ctx context.Context, conds ...NoteCondition) ([]*notedao.NotePO, error) {
	return d.repo.List(ctx, conds...)
}

// ListByCursor 游标分页查询笔记（支持 conditions）
func (d *NoteData) ListByCursor(ctx context.Context, cursor int64, limit int32, conds ...NoteCondition) ([]*notedao.NotePO, error) {
	return d.repo.ListByCursor(ctx, cursor, limit, conds...)
}

// ListByPage 页码分页查询笔记（支持 conditions）
func (d *NoteData) ListByPage(ctx context.Context, page, count int32, conds ...NoteCondition) ([]*notedao.NotePO, error) {
	return d.repo.ListByPage(ctx, page, count, conds...)
}

// Count 统计笔记数量（支持 conditions）
func (d *NoteData) Count(ctx context.Context, conds ...NoteCondition) (int64, error) {
	return d.repo.Count(ctx, conds...)
}

// GetRecentPublicPosted 获取用户最近发布的公开笔记
func (d *NoteData) GetRecentPublicPosted(ctx context.Context, uid int64, count int32) ([]*notedao.NotePO, error) {
	return d.repo.ListByPage(ctx, 1, count,
		WithNoteOwnerEqual(uid),
		WithNotePrivacyEqual(model.PrivacyPublic),
		WithNoteStateEqual(model.NoteStatePublished),
	)
}

// GetPublicByCursor 按游标获取公开笔记
func (d *NoteData) GetPublicByCursor(ctx context.Context, id int64, count int) ([]*notedao.NotePO, error) {
	return d.repo.GetPublicByCursor(ctx, id, count)
}

// GetPublicLastId 获取最后一个公开笔记ID
func (d *NoteData) GetPublicLastId(ctx context.Context) (int64, error) {
	return d.repo.GetPublicLastId(ctx)
}

// GetPublicAll 获取所有公开笔记
func (d *NoteData) GetPublicAll(ctx context.Context) ([]*notedao.NotePO, error) {
	return d.repo.List(ctx,
		WithNotePrivacyEqual(model.PrivacyPublic),
		WithNoteStateEqual(model.NoteStatePublished),
	)
}

// GetPublicCount 获取公开笔记数量
func (d *NoteData) GetPublicCount(ctx context.Context) (int64, error) {
	return d.repo.Count(ctx,
		WithNotePrivacyEqual(model.PrivacyPublic),
		WithNoteStateEqual(model.NoteStatePublished),
	)
}

// ConvertNotes 将dao层的NotePO转换为model层的Note
func (d *NoteData) ConvertNotes(notes []*notedao.NotePO) []*model.Note {
	result := make([]*model.Note, 0, len(notes))
	for _, n := range notes {
		result = append(result, &model.Note{
			NoteId:   n.Id,
			Title:    n.Title,
			Desc:     n.Desc,
			Privacy:  n.Privacy,
			Owner:    n.Owner,
			Type:     n.NoteType,
			State:    n.State,
			CreateAt: n.CreateAt,
			UpdateAt: n.UpdateAt,
		})
	}
	return result
}
