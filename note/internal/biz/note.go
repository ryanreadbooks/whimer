package biz

import (
	"context"
	"errors"

	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/oss"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// NoteBiz作为最基础的biz可以被其它biz依赖，其它biz之间不能相互依赖
type NoteBiz interface {
	// 获取笔记基础信息
	GetNote(ctx context.Context, noteId uint64) (*model.Note, error)
	// 判断笔记是否存在
	IsNoteExist(ctx context.Context, noteId uint64) (bool, error)
	// 获取笔记作者
	GetNoteOwner(ctx context.Context, noteId uint64) (uint64, error)
	// 组装笔记信息，主要是填充asset
	AssembleNotes(ctx context.Context, notes []*model.Note) (*model.Notes, error)
}

type noteBiz struct {
}

func NewNoteBiz() NoteBiz {
	b := &noteBiz{}
	return b
}

func (b *noteBiz) GetNote(ctx context.Context, noteId uint64) (*model.Note, error) {
	note, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz find one note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, model.NoteFromDao(note).AsSlice())
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.Items[0], nil
}

func (b *noteBiz) IsNoteExist(ctx context.Context, noteId uint64) (bool, error) {
	if noteId <= 0 {
		return false, nil
	}

	_, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return false, xerror.Wrapf(err, "note repo find one failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}
		return false, nil
	}

	return true, nil
}

func (b *noteBiz) GetNoteOwner(ctx context.Context, noteId uint64) (uint64, error) {
	n, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return 0, xerror.Wrapf(err, "biz find one failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}
		return 0, global.ErrNoteNotFound
	}

	return n.Owner, nil
}

// 组装笔记信息
func (b *noteBiz) AssembleNotes(ctx context.Context, notes []*model.Note) (*model.Notes, error) {
	var noteIds = make([]uint64, 0, len(notes))
	likesReq := make([]*counterv1.GetSummaryRequest, 0, len(notes))
	for _, note := range notes {
		noteIds = append(noteIds, note.NoteId)
		likesReq = append(likesReq, &counterv1.GetSummaryRequest{
			BizCode: global.NoteLikeBizcode,
			Oid:     note.NoteId,
		})
	}

	// 获取资源信息
	noteAssets, err := infra.Dao().NoteAssetRepo.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		return nil, xerror.Wrapf(err, "repo note asset failed")
	}

	// 组合notes和noteAssets
	var res model.Notes
	for _, note := range notes {
		item := &model.Note{
			NoteId:   note.NoteId,
			Title:    note.Title,
			Desc:     note.Desc,
			Privacy:  note.Privacy,
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		for _, asset := range noteAssets {
			if note.NoteId == asset.NoteId {
				item.Images = append(item.Images, &model.NoteImage{
					Url: oss.GetPublicVisitUrl(
						config.Conf.Oss.Bucket,
						asset.AssetKey,
						config.Conf.Oss.DisplayEndpoint,
					),
					Type: int(asset.AssetType),
				})
			}
		}

		res.Items = append(res.Items, item)
	}

	return &res, nil
}
