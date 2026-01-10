package data

import (
	"context"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
)

// NoteExtData 笔记扩展信息数据层
type NoteExtData struct {
	repo *notedao.NoteExtRepo
}

func NewNoteExtData(repo *notedao.NoteExtRepo) *NoteExtData {
	return &NoteExtData{
		repo: repo,
	}
}

// Upsert 插入或更新扩展信息
func (d *NoteExtData) Upsert(ctx context.Context, ext *notedao.ExtPO) error {
	return d.repo.Upsert(ctx, ext)
}

// Delete 删除扩展信息
func (d *NoteExtData) Delete(ctx context.Context, noteId int64) error {
	return d.repo.Delete(ctx, noteId)
}

// GetById 根据笔记ID获取扩展信息
func (d *NoteExtData) GetById(ctx context.Context, noteId int64) (*notedao.ExtPO, error) {
	return d.repo.GetById(ctx, noteId)
}

// BatchGetById 批量获取扩展信息
func (d *NoteExtData) BatchGetById(ctx context.Context, noteIds []int64) ([]*notedao.ExtPO, error) {
	return d.repo.BatchGetById(ctx, noteIds)
}
