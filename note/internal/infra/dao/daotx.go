package dao

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 多表涉及事务的操作

// 事务插入一条笔记
func (d *Dao) CreateNote(ctx context.Context, note *Note, assets []*NoteAsset) (uint64, error) {
	var noteId uint64
	err := d.db.TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 插入图片基础内容
		var errTx error
		noteId, errTx = d.NoteDao.InsertTx(ctx, tx, note)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note dao insert tx failed")
		}

		// 填充noteId
		for _, a := range assets {
			a.NoteId = noteId
		}

		// 插入笔记资源数据
		errTx = d.NoteAssetRepo.BatchInsertTx(ctx, tx, assets)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note asset dao batch insert tx failed")
		}

		return nil
	})

	if err != nil {
		return 0, xerror.Wrapf(err, "dao transact insert note failed")
	}

	return noteId, nil
}

// 事务更新一条笔记，包括更新笔记基础信息和笔记资源
func (d *Dao) UpdateNote(ctx context.Context, note *Note, assets []*NoteAsset) error {
	var now = time.Now().Unix()

	err := d.db.TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 先更新基础信息
		err := d.NoteDao.UpdateTx(ctx, tx, note)
		if err != nil {
			return xerror.Wrapf(err, "note dao update tx failed")
		}

		// 找出旧资源
		oldAssets, err := d.NoteAssetRepo.FindImageByNoteIdTx(ctx, tx, note.Id)
		if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
			return xerror.Wrapf(err, "noteasset dao find failed")
		}

		// 笔记的新资源
		newAssetKeys := make([]string, 0, len(assets))
		for _, asset := range assets {
			newAssetKeys = append(newAssetKeys, asset.AssetKey)
		}

		// 随后删除旧资源
		// 删除除了newAssetKeys之外的其它
		err = d.NoteAssetRepo.ExcludeDeleteImageByNoteIdTx(ctx, tx, note.Id, newAssetKeys)
		if err != nil {
			return xerror.Wrapf(err, "noteasset dao delete tx failed")
		}

		// 找出old和new的资源差异，只更新发生了变化的部分
		oldAssetMap := make(map[string]struct{})
		for _, old := range oldAssets {
			oldAssetMap[old.AssetKey] = struct{}{}
		}
		newAssets := make([]*NoteAsset, 0, len(assets))
		for _, asset := range assets {
			if _, ok := oldAssetMap[asset.AssetKey]; !ok {
				newAssets = append(newAssets, &NoteAsset{
					AssetKey:  asset.AssetKey,
					AssetType: global.AssetTypeImage,
					NoteId:    note.Id,
					CreateAt:  now,
				})
			}
		}

		if len(newAssets) == 0 {
			return nil
		}

		// 插入新的资源
		err = d.NoteAssetRepo.BatchInsertTx(ctx, tx, newAssets)
		if err != nil {
			return xerror.Wrapf(err, "noteasset dao batch insert tx failed")
		}

		return nil
	})

	return xerror.Wrapf(err, "dao transact update note failed")
}

// 事务删除一条笔记，包括删除笔记基本信息和笔记资源
func (d *Dao) DeleteNote(ctx context.Context, noteId uint64) error {
	err := d.db.TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		err := d.NoteDao.DeleteTx(ctx, tx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "dao delete note basic tx failed")
		}

		err = d.NoteAssetRepo.DeleteByNoteIdTx(ctx, tx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "dao delete note asset tx failed")
		}

		return nil
	})

	return xerror.Wrapf(err, "dao transact delete note failed")
}
