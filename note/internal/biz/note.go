package biz

import (
	"context"
	"errors"
	"math"

	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

// NoteBiz作为最基础的biz可以被其它biz依赖，其它biz之间不能相互依赖
type NoteBiz struct {
}

func NewNoteBiz() NoteBiz {
	b := NoteBiz{}
	return b
}

// 获取笔记基础信息
func (b *NoteBiz) GetNote(ctx context.Context, noteId uint64) (*model.Note, error) {
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

// 获取用户最近发布的笔记
func (b *NoteBiz) GetUserRecentNote(ctx context.Context, uid int64, count int32) (*model.Notes, error) {
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

func (b *NoteBiz) ListUserPublicNote(ctx context.Context, uid int64, cursor uint64, count int32) (*model.Notes, model.PageResult, error) {
	var pageResult = model.PageResult{}
	if cursor == 0 {
		cursor = math.MaxUint64
	}

	notes, err := infra.Dao().NoteDao.ListPublicByOwnerByCursor(ctx, uid, cursor, count)
	if err != nil {
		if xsql.IsNotFound(err) {
			return &model.Notes{}, pageResult, nil
		}

		return nil, pageResult, xerror.Wrapf(err, "biz list notes failed").WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
	if err != nil {
		return nil, pageResult, xerror.Wrapf(err, "biz assemble notes failed when get recent notes").
			WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	// 计算下一次请求的游标位置
	if len(notes) > 0 {
		pageResult.NextCursor = notes[len(notes)-1].Id
		if len(notes) == int(count) {
			pageResult.HasNext = true
		}
	}

	return resp, pageResult, nil
}

func (b *NoteBiz) GetPublicNote(ctx context.Context, noteId uint64) (*model.Note, error) {
	note, err := b.GetNote(ctx, noteId)
	if err != nil {
		return nil, err
	}

	if note.Privacy != global.PrivacyPublic {
		return nil, global.ErrNoteNotPublic
	}

	return note, nil
}

// 判断笔记是否存在
func (b *NoteBiz) IsNoteExist(ctx context.Context, noteId uint64) (bool, error) {
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

// 获取笔记作者
func (b *NoteBiz) GetNoteOwner(ctx context.Context, noteId uint64) (int64, error) {
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
func (b *NoteBiz) AssembleNotes(ctx context.Context, notes []*model.Note) (*model.Notes, error) {
	var noteIds = make([]uint64, 0, len(notes))
	for _, n := range notes {
		noteIds = append(noteIds, n.NoteId)
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

		k, s := config.Conf.ImgProxyAuth.GetKey(), config.Conf.ImgProxyAuth.GetSalt()
		for _, asset := range noteAssets {
			assetMeta := model.NewAssetImageMetaFromJson(asset.AssetMeta)
			if note.NoteId == asset.NoteId {
				// pureKey := strings.TrimLeft(asset.AssetKey, config.Conf.Oss.Bucket+"/") // 此处要去掉桶名称
				item.Images = append(item.Images, &model.NoteImage{
					// TODO 大图片的占用还是太大了
					Url:    imgproxy.GetSignedUrl(config.Conf.Oss.DisplayEndpointBucket(), asset.AssetKey, k, s, imgproxy.WithQuality("15")),
					UrlPrv: imgproxy.GetSignedUrl(config.Conf.Oss.DisplayEndpointBucket(), asset.AssetKey, k, s, imgproxy.WithQuality("1")),
					Type:   int(asset.AssetType),
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
