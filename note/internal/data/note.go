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
	return d.repo.Insert(ctx, note)
}

// Update 更新笔记，同时删除相关缓存
func (d *NoteData) Update(ctx context.Context, note *notedao.NotePO) error {
	err := d.repo.Update(ctx, note)
	if err != nil {
		return err
	}

	// 异步删除缓存
	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := d.cache.DelNoteRelatedCache(ctx, note.Id, note.Owner); err != nil {
			xlog.Msg("note data failed to del cache when updating").
				Extras("noteId", note.Id).Err(err).Errorx(ctx)
		}
	})

	return nil
}

// Delete 删除笔记，同时删除相关缓存
func (d *NoteData) Delete(ctx context.Context, noteId int64) error {
	// 先获取owner用于删除缓存
	ownerId, err := d.repo.GetOwner(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "get owner failed")
	}

	// 删除笔记
	err = d.repo.Delete(ctx, noteId)
	if err != nil {
		return err
	}

	// 异步删除缓存
	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := d.cache.DelNoteRelatedCache(ctx, noteId, ownerId); err != nil {
			xlog.Msg("note data failed to del cache when deleting").
				Extras("noteId", noteId).Err(err).Errorx(ctx)
		}
	})

	return nil
}

// UpdateState 更新笔记状态，同时删除缓存
func (d *NoteData) UpdateState(ctx context.Context, noteId int64, state model.NoteState) error {
	err := d.repo.UpdateState(ctx, noteId, state)
	if err != nil {
		return err
	}

	// 异步删除缓存
	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := d.cache.DelNote(ctx, noteId); err != nil {
			xlog.Msg("note data failed to del cache when updating state").
				Extras("noteId", noteId).Err(err).Errorx(ctx)
		}
	})

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
	cnt, err := d.repo.GetPostedCountByOwner(ctx, uid)
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
	cnt, err := d.repo.GetPublicPostedCountByOwner(ctx, uid)
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

// ListByOwner 获取用户的所有笔记
func (d *NoteData) ListByOwner(ctx context.Context, uid int64) ([]*notedao.NotePO, error) {
	return d.repo.ListByOwner(ctx, uid)
}

// ListByOwnerByCursor 分页获取用户的笔记
func (d *NoteData) ListByOwnerByCursor(ctx context.Context, uid int64, cursor int64, limit int32) ([]*notedao.NotePO, error) {
	return d.repo.ListByOwnerByCursor(ctx, uid, cursor, limit)
}

// ListPublicByOwnerByCursor 分页获取用户的公开笔记
func (d *NoteData) ListPublicByOwnerByCursor(ctx context.Context, uid int64, cursor int64, limit int32) ([]*notedao.NotePO, error) {
	return d.repo.ListPublicByOwnerByCursor(ctx, uid, cursor, limit)
}

// ListPublicByCursor 分页获取公开笔记
func (d *NoteData) ListPublicByCursor(ctx context.Context, cursor int64, limit int32) ([]*notedao.NotePO, error) {
	return d.repo.ListPublicByCursor(ctx, cursor, limit)
}

// PageListByOwner 分页获取用户的笔记
func (d *NoteData) PageListByOwner(ctx context.Context, uid int64, page, count int32) ([]*notedao.NotePO, error) {
	return d.repo.PageListByOwner(ctx, uid, page, count)
}

// GetRecentPublicPosted 获取用户最近发布的公开笔记
func (d *NoteData) GetRecentPublicPosted(ctx context.Context, uid int64, count int32) ([]*notedao.NotePO, error) {
	return d.repo.GetRecentPublicPosted(ctx, uid, count)
}

// GetPublicByCursor 按游标获取公开笔记
func (d *NoteData) GetPublicByCursor(ctx context.Context, id int64, count int) ([]*notedao.NotePO, error) {
	return d.repo.GetPublicByCursor(ctx, id, count)
}

// GetPrivateByCursor 按游标获取私有笔记
func (d *NoteData) GetPrivateByCursor(ctx context.Context, id int64, count int) ([]*notedao.NotePO, error) {
	return d.repo.GetPrivateByCursor(ctx, id, count)
}

// GetPublicLastId 获取最后一个公开笔记ID
func (d *NoteData) GetPublicLastId(ctx context.Context) (int64, error) {
	return d.repo.GetPublicLastId(ctx)
}

// GetPrivateLastId 获取最后一个私有笔记ID
func (d *NoteData) GetPrivateLastId(ctx context.Context) (int64, error) {
	return d.repo.GetPrivateLastId(ctx)
}

// GetPublicAll 获取所有公开笔记
func (d *NoteData) GetPublicAll(ctx context.Context) ([]*notedao.NotePO, error) {
	return d.repo.GetPublicAll(ctx)
}

// GetPrivateAll 获取所有私有笔记
func (d *NoteData) GetPrivateAll(ctx context.Context) ([]*notedao.NotePO, error) {
	return d.repo.GetPrivateAll(ctx)
}

// GetPublicCount 获取公开笔记数量
func (d *NoteData) GetPublicCount(ctx context.Context) (int64, error) {
	return d.repo.GetPublicCount(ctx)
}

// GetPrivateCount 获取私有笔记数量
func (d *NoteData) GetPrivateCount(ctx context.Context) (int64, error) {
	return d.repo.GetPrivateCount(ctx)
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
