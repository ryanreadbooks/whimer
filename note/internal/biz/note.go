package biz

import (
	"context"
	"errors"
	"strings"

	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/oss"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	"github.com/ryanreadbooks/whimer/asset-job/sdk/rules"
)

// NoteBiz作为最基础的biz可以被其它biz依赖，其它biz之间不能相互依赖
type NoteBiz interface {
	// 获取笔记基础信息
	GetNote(ctx context.Context, noteId uint64) (*model.Note, error)
	// 获取用户最近发布的笔记
	GetUserRecentNote(ctx context.Context, uid uint64, count int32) (*model.Notes, error)
	// 获取公开的笔记基础信息
	GetPublicNote(ctx context.Context, noteId uint64) (*model.Note, error)
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
		if xsql.IsNotFound(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz find one note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, model.NoteFromDao(note).AsSlice())
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.Items[0], nil
}

func (b *noteBiz) GetUserRecentNote(ctx context.Context, uid uint64, count int32) (*model.Notes, error) {
	notes, err := infra.Dao().NoteDao.GetRecentPublicPosted(ctx, uid, count)
	if err != nil {
		if xsql.IsNotFound(err) {
			return &model.Notes{}, nil
		}

		return nil, xerror.Wrapf(err, "biz find recent posted failed").WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed when get recent notes").
			WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	return resp, nil
}

func (b *noteBiz) GetPublicNote(ctx context.Context, noteId uint64) (*model.Note, error) {
	note, err := b.GetNote(ctx, noteId)
	if err != nil {
		return nil, err
	}

	if note.Privacy != global.PrivacyPublic {
		return nil, global.ErrNoteNotPublic
	}

	return note, nil
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
// 笔记的资源数据，点赞等
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
			Owner:    note.Owner,
		}
		for _, asset := range noteAssets {
			assetMeta := model.NewAssetImageMetaFromJson(asset.AssetMeta)
			if note.NoteId == asset.NoteId {
				pureKey := strings.TrimLeft(asset.AssetKey, config.Conf.Oss.Bucket+"/") // 此处要去掉桶名称
				item.Images = append(item.Images, &model.NoteImage{
					Url: oss.GetPublicVisitUrl2(
						asset.AssetKey,
						config.Conf.Oss.DisplayEndpoint,
					),
					UrlPrv: oss.GetPublicVisitUrl(
						config.Conf.Oss.BucketPreview,
						rules.PreviewKey(pureKey),
						config.Conf.Oss.DisplayEndpoint,
					),
					Type: int(asset.AssetType),
					Meta: model.NoteImageMeta{
						Width:  assetMeta.Width,
						Height: assetMeta.Height,
						Format: assetMeta.Format,
					},
				})
			}
		}

		res.Items = append(res.Items, item)
	}

	return &res, nil
}
