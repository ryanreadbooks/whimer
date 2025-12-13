package data

import (
	"context"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
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

// FindOne 获取单个资源
func (d *NoteAssetData) FindOne(ctx context.Context, id int64) (*notedao.AssetPO, error) {
	return d.repo.FindOne(ctx, id)
}

// Insert 插入资源
func (d *NoteAssetData) Insert(ctx context.Context, asset *notedao.AssetPO) error {
	return d.repo.Insert(ctx, asset)
}

// BatchInsert 批量插入资源
func (d *NoteAssetData) BatchInsert(ctx context.Context, assets []*notedao.AssetPO) error {
	return d.repo.BatchInsert(ctx, assets)
}

// FindByNoteIds 根据笔记ID批量获取资源
func (d *NoteAssetData) FindByNoteIds(ctx context.Context, noteIds []int64) ([]*notedao.AssetPO, error) {
	return d.repo.FindByNoteIds(ctx, noteIds)
}

// FindImageByNoteId 获取笔记的图片资源
func (d *NoteAssetData) FindImageByNoteId(ctx context.Context, noteId int64) ([]*notedao.AssetPO, error) {
	return d.repo.FindImageByNoteId(ctx, noteId)
}

// DeleteByNoteId 删除笔记的所有资源
func (d *NoteAssetData) DeleteByNoteId(ctx context.Context, noteId int64) error {
	return d.repo.DeleteByNoteId(ctx, noteId)
}

// ExcludeDeleteImageByNoteId 删除笔记的图片资源（排除指定的key）
func (d *NoteAssetData) ExcludeDeleteImageByNoteId(ctx context.Context, noteId int64, assetKeys []string) error {
	return d.repo.ExcludeDeleteImageByNoteId(ctx, noteId, assetKeys)
}
