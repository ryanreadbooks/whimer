package data

import (
	"context"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// NoteAssetData 笔记资源数据层
type NoteAssetData struct {
	repo *notedao.NoteAssetRepo
}

func NewNoteAssetData(repo *notedao.NoteAssetRepo) *NoteAssetData {
	return &NoteAssetData{
		repo: repo,
	}
}

// 获取单个资源
func (d *NoteAssetData) FindOne(ctx context.Context, id int64) (*notedao.AssetPO, error) {
	return d.repo.FindOne(ctx, id)
}

// 插入资源
func (d *NoteAssetData) Insert(ctx context.Context, asset *notedao.AssetPO) error {
	return d.repo.Insert(ctx, asset)
}

// 批量插入资源
func (d *NoteAssetData) BatchInsert(ctx context.Context, assets []*notedao.AssetPO) error {
	return d.repo.BatchInsert(ctx, assets)
}

// 根据笔记ID批量获取资源
func (d *NoteAssetData) FindByNoteIds(ctx context.Context, noteIds []int64) ([]*notedao.AssetPO, error) {
	return d.repo.FindByNoteIds(ctx, noteIds)
}

// 获取笔记的图片资源
func (d *NoteAssetData) FindImageNoteAssets(ctx context.Context, noteId int64) ([]*notedao.AssetPO, error) {
	return d.repo.FindByNoteIdForUpdate(ctx, noteId, model.AssetTypeImage)
}

func (d *NoteAssetData) FindVideoNoteAssets(ctx context.Context, noteId int64) ([]*notedao.AssetPO, error) {
	return d.repo.FindByNoteIdForUpdate(ctx, noteId, model.AssetTypeVideo, model.AssetTypeImage)
}

// 删除笔记的所有资源
func (d *NoteAssetData) DeleteByNoteId(ctx context.Context, noteId int64) error {
	return d.repo.DeleteByNoteId(ctx, noteId)
}

// 删除笔记的图片资源（排除指定的key）
func (d *NoteAssetData) DeleteImageByNoteIdExcept(ctx context.Context, noteId int64, assetKeys []string) error {
	return d.repo.ExcludeDeleteByNoteId(ctx, noteId, assetKeys, model.AssetTypeImage)
}

// 批量更新资源元数据
func (d *NoteAssetData) BatchUpdateAssetMeta(ctx context.Context, noteId int64, metas map[string][]byte) error {
	return d.repo.BatchUpdateAssetMeta(ctx, noteId, metas)
}
