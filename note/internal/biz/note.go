package biz

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	tagdao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/tag"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/model/convert"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

// NoteBiz作为最基础的biz可以被其它biz依赖，其它biz之间不能相互依赖
type NoteBiz struct {
	data *data.Data
}

func NewNoteBiz(dt *data.Data) *NoteBiz {
	return &NoteBiz{
		data: dt,
	}
}

func (b *NoteBiz) GetNoteType(ctx context.Context, noteId int64) (model.NoteType, error) {
	noteType, err := b.data.Note.GetNoteType(ctx, noteId)
	if err != nil {
		return 0, xerror.Wrapf(err, "biz get note type failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return noteType, nil
}

func (b *NoteBiz) GetNoteWithoutCache(ctx context.Context, noteId int64) (*model.Note, error) {
	note, err := b.data.Note.FindOne(ctx, noteId, data.WithoutCache())
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz find one note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return convert.NoteFromDao(note), nil
}

// 获取笔记基础信息
// 不包含点赞等互动信息和标签
func (b *NoteBiz) GetNote(ctx context.Context, noteId int64) (*model.Note, error) {
	note, err := b.data.Note.FindOne(ctx, noteId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz find one note failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, convert.NoteFromDao(note).AsSlice())
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return resp.Items[0], nil
}

// 批量获取笔记基础信息 不包含点赞等互动信息和标签
func (b *NoteBiz) BatchGetNote(ctx context.Context, noteIds []int64) (map[int64]*model.Note, error) {
	notesMap, err := b.data.Note.BatchGet(ctx, noteIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz batch get note failed").WithCtx(ctx)
	}

	if len(notesMap) == 0 {
		return map[int64]*model.Note{}, nil
	}

	notes := xmap.Values(notesMap)
	assembledNotes, err := b.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed").WithCtx(ctx)
	}

	resp := make(map[int64]*model.Note, len(notes))
	for _, n := range assembledNotes.Items {
		resp[n.NoteId] = n
	}

	return resp, nil
}

// 批量获取笔记基础信息 不包含点赞等互动信息和标签 不包含asset
func (b *NoteBiz) BatchGetNoteWithoutAsset(ctx context.Context, noteIds []int64) (map[int64]*model.Note, error) {
	notesMap, err := b.data.Note.BatchGet(ctx, noteIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz batch get note failed").WithCtx(ctx)
	}

	if len(notesMap) == 0 {
		return map[int64]*model.Note{}, nil
	}

	daoNotes := xmap.Values(notesMap)
	notes := convert.NoteSliceFromDao(daoNotes)

	resp := make(map[int64]*model.Note, len(notes))
	for _, n := range notes {
		resp[n.NoteId] = n
	}

	return resp, nil
}

// 获取用户最近发布的笔记
func (b *NoteBiz) GetUserRecentNote(ctx context.Context, uid int64, count int32) (*model.Notes, error) {
	notes, err := b.data.Note.GetRecentPublicPosted(ctx, uid, count)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return &model.Notes{}, nil
		}

		return nil, xerror.Wrapf(err, "biz find recent posted failed").WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	resp, err := b.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil, xerror.Wrapf(err, "biz assemble notes failed when get recent notes").
			WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	return resp, nil
}

func (b *NoteBiz) ListUserPublicNote(ctx context.Context, uid int64, cursor int64, count int32) (*model.Notes, model.PageResult, error) {
	nextPage := model.PageResult{}
	if cursor == 0 {
		cursor = model.MaxCursor
	}

	newCount := count + 1

	notes, err := b.data.Note.ListByCursor(ctx, cursor, newCount,
		data.WithNoteOwnerEqual(uid),
		data.WithNotePrivacyEqual(model.PrivacyPublic),
		data.WithNoteStateEqual(model.NoteStatePublished),
	)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return &model.Notes{}, nextPage, nil
		}

		return nil, nextPage, xerror.Wrapf(err, "biz list notes failed").WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	gotLen := len(notes)
	if gotLen == int(newCount) {
		// has more
		notes = notes[0 : gotLen-1]
		// 计算下一次请求的游标位置
		nextPage.HasNext = true
		nextPage.NextCursor = notes[len(notes)-1].Id
	} else {
		nextPage.HasNext = false
	}

	resp, err := b.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil,
			nextPage,
			xerror.Wrapf(err, "biz assemble notes failed when get recent notes").
				WithExtras("uid", uid, "count", count).WithCtx(ctx)
	}

	return resp, nextPage, nil
}

func (b *NoteBiz) GetPublicNote(ctx context.Context, noteId int64) (*model.Note, error) {
	note, err := b.GetNote(ctx, noteId)
	if err != nil {
		return nil, err
	}

	if note.Privacy != model.PrivacyPublic {
		return nil, global.ErrNoteNotPublic
	}

	return note, nil
}

// 判断笔记是否存在
func (b *NoteBiz) IsNoteExist(ctx context.Context, noteId int64) (bool, error) {
	if noteId <= 0 {
		return false, nil
	}

	_, err := b.data.Note.FindOne(ctx, noteId)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return false, xerror.Wrapf(err, "note repo find one failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}
		return false, nil
	}

	return true, nil
}

// 获取笔记作者
func (b *NoteBiz) GetNoteOwner(ctx context.Context, noteId int64) (int64, error) {
	n, err := b.data.Note.FindOne(ctx, noteId)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return 0, xerror.Wrapf(err, "biz find one failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}
		return 0, global.ErrNoteNotFound
	}

	return n.Owner, nil
}

// 组装笔记信息 笔记的资源数据
func (b *NoteBiz) AssembleNotes(ctx context.Context, notes []*model.Note) (*model.Notes, error) {
	var res model.Notes

	noteIds := make([]int64, 0, len(notes))
	for _, n := range notes {
		noteIds = append(noteIds, n.NoteId)
	}

	// 获取资源信息
	noteAssets, err := b.data.NoteAsset.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		return nil, xerror.Wrapf(err, "repo note asset failed")
	}

	// 组合notes和noteAssets
	for _, note := range notes {
		if note.Type == model.AssetTypeVideo {
			note.Videos = &model.NoteVideo{}
		}

		for _, asset := range noteAssets {
			switch asset.AssetType {
			case model.AssetTypeImage:
				assetMeta := model.NewAssetImageMetaFromJson(asset.AssetMeta)
				if note.NoteId == asset.NoteId {
					// pureKey := strings.TrimLeft(asset.AssetKey, config.Conf.Oss.Bucket+"/") // 此处要去掉桶名称
					// 放在pilot服务处理proxy
					note.Images = append(note.Images, &model.NoteImage{
						Key:  asset.AssetKey,
						Type: int(asset.AssetType),
						Meta: model.NoteImageMeta{
							Width:  assetMeta.Width,
							Height: assetMeta.Height,
							Format: assetMeta.Format,
						},
					})
				}
			case model.AssetTypeVideo:
				videoMeta := model.NewVideoInfoFromJson(asset.AssetMeta)
				if note.NoteId == asset.NoteId {
					note.Videos.Items = append(note.Videos.Items, &model.NoteVideoItem{
						Key: asset.AssetKey,
						Media: &model.NoteVideoMedia{
							Width:        videoMeta.Width,
							Height:       videoMeta.Height,
							VideoCodec:   videoMeta.Codec,
							Bitrate:      videoMeta.Bitrate,
							FrameRate:    videoMeta.Framerate,
							Duration:     videoMeta.Duration,
							Format:       "video/mp4",
							AudioCodec:   videoMeta.AudioCodec,
							AudioBitrate: videoMeta.AudioBitrate,
						},
					})
				}
			}
		}

		res.Items = append(res.Items, note)
	}

	return &res, nil
}

// 按需填充ext所属内容
func (b *NoteBiz) AssembleNotesExt(ctx context.Context, notes []*model.Note) error {
	if len(notes) == 0 {
		return nil
	}

	noteIds := make([]int64, 0, len(notes))
	for _, n := range notes {
		noteIds = append(noteIds, n.NoteId)
	}

	noteIds = xslice.Uniq(noteIds)
	extsPo, err := b.data.NoteExt.BatchGetById(ctx, noteIds)
	if err != nil {
		return xerror.Wrapf(err, "note ext dao failed to batch get")
	}

	// noteId -> ext
	extsPoMap := xslice.MakeMap(extsPo, func(e *notedao.ExtPO) int64 { return e.NoteId })
	extMap := make(map[int64]*model.NoteExt, len(extsPoMap))
	totalTagIds := make([]int64, 0, len(extsPoMap))

	for noteId, extPo := range extsPoMap {
		if extPo == nil {
			continue
		}

		tIds := xslice.SplitInts[int64](extPo.Tags, ",")
		totalTagIds = append(totalTagIds, tIds...)
		extMap[noteId] = &model.NoteExt{}
		extMap[noteId].TagIds = tIds

		if extPo.AtUsers != nil {
			var atUsers []*model.AtUser
			if errParseAtUsers := json.Unmarshal(extPo.AtUsers, &atUsers); errParseAtUsers == nil {
				extMap[noteId].AtUsers = atUsers
			}
		}
	}

	totalTagIds = xslice.Uniq(totalTagIds)
	tagMap := make(map[int64]*tagdao.Tag)
	// query tags
	if len(totalTagIds) != 0 {
		tags, err := b.data.Tag.BatchGetById(ctx, totalTagIds)
		if err != nil {
			return xerror.Wrapf(err, "tag dao failed to batch get")
		}
		tagMap = xslice.MakeMap(tags, func(e *tagdao.Tag) int64 { return e.Id })
	}

	// do the assignment
	for _, n := range notes {
		if ext, ok := extMap[n.NoteId]; ok && ext != nil {
			// 1. tags
			for _, tagId := range ext.TagIds {
				if tag, tagOk := tagMap[tagId]; tagOk {
					n.Tags = append(n.Tags, &model.NoteTag{
						Id:    tag.Id,
						Name:  tag.Name,
						Ctime: tag.Ctime,
					})
				}
			}

			// 2. at_users
			n.AtUsers = ext.AtUsers

			// TODO other ext attributes if necessary
		}
	}

	return nil
}

func (b *NoteBiz) GetTag(ctx context.Context, tagId int64) (*model.NoteTag, error) {
	tag, err := b.data.Tag.FindById(ctx, tagId)
	if err != nil {
		return nil, xerror.Wrapf(err, "tag dao failed to get").WithExtra("tag_id", tagId).WithCtx(ctx)
	}

	return convert.NoteTagFromDao(tag), nil
}

func (b *NoteBiz) BatchGetTag(ctx context.Context, tagIds []int64) (map[int64]*model.NoteTag, error) {
	tags, err := b.data.Tag.BatchGetById(ctx, tagIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "tag dao failed to batch get").WithExtra("tag_ids", tagIds).WithCtx(ctx)
	}

	res := make(map[int64]*model.NoteTag, len(tags))
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		res[tag.Id] = convert.NoteTagFromDao(tag)
	}

	return res, nil
}
