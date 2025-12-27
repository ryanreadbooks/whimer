package biz

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	tagdao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/tag"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/model/convert"
)

// 笔记相关
// 创作者相关
type NoteCreatorBiz struct {
	data *data.Data
	note *NoteBiz
}

func NewNoteCreatorBiz(dt *data.Data, note *NoteBiz) *NoteCreatorBiz {
	return &NoteCreatorBiz{
		data: dt,
		note: note,
	}
}

func isNoteExtValid(ext *notedao.ExtPO) bool {
	if ext == nil {
		return false
	}

	if ext.Tags != "" || len(ext.AtUsers) > 0 {
		return true
	}

	return false
}

func assignNoteAssets(newNote *model.Note,
	req *CreateNoteRequest,
) []*notedao.AssetPO {
	now := time.Now().Unix()
	var noteAssets []*notedao.AssetPO
	switch req.Basic.NoteType {
	case model.AssetTypeImage:
		noteAssets = make([]*notedao.AssetPO, 0, len(req.Images))
		for _, img := range req.Images {
			imgMeta := model.NewAssetImageMeta(img.Width, img.Height, img.Format).Bytes()
			noteAssets = append(noteAssets, &notedao.AssetPO{
				NoteId:    newNote.NoteId,
				AssetKey:  img.FileId,           // 包含桶名称
				AssetType: model.AssetTypeImage, // image
				CreateAt:  now,
				AssetMeta: imgMeta,
			})
			mmimg := &model.NoteImage{
				Key:  img.FileId,
				Type: int(model.AssetTypeImage),
				Meta: model.NoteImageMeta{
					Width:  img.Width,
					Height: img.Height,
					Format: img.Format,
				},
			}
			mmimg.SetBucket(img.Bucket)
			newNote.Images = append(newNote.Images, mmimg)
		}
	case model.AssetTypeVideo:
		noteAssets = formatNoteVideoAsset(req.Video)
		for _, asset := range noteAssets {
			asset.NoteId = newNote.NoteId
		}
		items := make([]*model.NoteVideoItem, 0, len(noteAssets))
		for _, asset := range noteAssets {
			items = append(items, &model.NoteVideoItem{
				Key:   asset.AssetKey,
				Media: &model.NoteVideoMedia{},
			})
		}
		newNote.Videos = &model.NoteVideo{
			Items: items,
		}
		newNote.Videos.SetTargetBucket(req.Video.TargetBucket)
		newNote.Videos.SetRawUrl(req.Video.FileId)
		newNote.Videos.SetRawBucket(req.Video.Bucket)

		// 加上视频封面
		coverImage := &model.AssetImageMeta{}
		noteAssets = append(noteAssets, &notedao.AssetPO{
			NoteId:    newNote.NoteId,
			AssetKey:  req.Video.CoverFileId,
			AssetType: model.AssetTypeImage,
			CreateAt:  now,
			AssetMeta: coverImage.Bytes(),
		})
	}

	return noteAssets
}

func (b *NoteCreatorBiz) CreateNote(ctx context.Context, req *CreateNoteRequest) (*model.Note, error) {
	var (
		uid    = metadata.Uid(ctx)
		ip     = xnet.IpAsBytes(metadata.ClientIp(ctx))
		noteId int64
	)

	now := time.Now().Unix()
	newNotePO := &notedao.NotePO{
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  req.Basic.Privacy,
		NoteType: req.Basic.NoteType,
		State:    model.NoteStateInit,
		Owner:    uid,
		Ip:       ip,
		CreateAt: now,
		UpdateAt: now,
	}

	newNote := convert.NoteFromDao(newNotePO)
	noteAssetsPO := assignNoteAssets(newNote, req)

	// begin tx
	err := b.data.DB().Transact(ctx, func(ctx context.Context) error {
		// 插入图片基础内容
		var errTx error
		noteId, errTx = b.data.Note.Insert(ctx, newNotePO)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note dao insert tx failed")
		}

		newNotePO.Id = noteId

		// 回填noteId
		for _, a := range noteAssetsPO {
			a.NoteId = noteId
		}
		// 插入笔记资源数据
		errTx = b.data.NoteAsset.BatchInsert(ctx, noteAssetsPO)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note asset dao batch insert tx failed")
		}

		// 笔记额外信息
		if errTx = b.upsertNoteExt(ctx, noteId, req); errTx != nil {
			return xerror.Wrapf(errTx, "note ext insert tx failed")
		}

		return nil
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "biz create note failed").WithExtra("note", req).WithCtx(ctx)
	}

	// 回填noteId
	newNote.NoteId = noteId

	return newNote, nil
}

// 保存笔记扩展信息 包括扩展信息和at人等额外信息
func (b *NoteCreatorBiz) upsertNoteExt(ctx context.Context, noteId int64, req *CreateNoteRequest) error {
	ext := &notedao.ExtPO{
		NoteId:  noteId,
		Tags:    xslice.JoinInts(req.TagIds),
		AtUsers: json.RawMessage{},
	}
	if len(req.AtUsers) > 0 {
		if data, err := json.Marshal(req.AtUsers); err == nil {
			ext.AtUsers = data
		}
	}
	if isNoteExtValid(ext) {
		return b.data.NoteExt.Upsert(ctx, ext)
	}

	return nil
}

// 更新笔记
func (b *NoteCreatorBiz) UpdateNote(ctx context.Context, req *UpdateNoteRequest) (*model.Note, error) {
	var (
		uid = metadata.Uid(ctx)
		ip  = xnet.IpAsBytes(metadata.ClientIp(ctx))
	)

	now := time.Now().Unix()
	noteId := req.NoteId
	oldNote, err := b.data.Note.FindOneForUpdate(ctx, noteId) // lock record for update
	if errors.Is(err, xsql.ErrNoRecord) {
		return nil, global.ErrNoteNotFound
	}
	if err != nil {
		return nil, xerror.Wrapf(err, "biz find one note failed").WithExtra("note", req).WithCtx(ctx)
	}

	if oldNote.NoteType != req.Basic.NoteType {
		return nil, global.ErrNoteTypeCannotChange
	}

	// 确保更新者uid和笔记作者uid相同
	if uid != oldNote.Owner {
		return nil, global.ErrNotNoteOwner
	}

	newNotePO := &notedao.NotePO{
		Id:       noteId,
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  req.Basic.Privacy,
		NoteType: oldNote.NoteType,
		State:    model.NoteStateInit, // 更新后需要重新走流程 ? 还是可以在特定阶段重新开始
		CreateAt: oldNote.CreateAt,
		UpdateAt: now,
		Owner:    oldNote.Owner,
		Ip:       ip,
	}
	newNote := convert.NoteFromDao(newNotePO)
	assetUpdated, err := b.handleNoteAssets(ctx, newNote, &req.CreateNoteRequest, true)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz update note assets failed")
	}
	if !assetUpdated {
		// 资源没有变 就不需要重新走资源处理流程 直接从审核阶段开始
		newNote.State = model.NoteStateAuditing
		newNotePO.State = model.NoteStateAuditing
	}

	// 更新基础信息
	err = b.data.Note.Update(ctx, newNotePO)
	if err != nil {
		return nil, xerror.Wrapf(err, "note dao update tx failed")
	}

	// ext处理
	if err = b.upsertNoteExt(ctx, oldNote.Id, &req.CreateNoteRequest); err != nil {
		return nil, xerror.Wrapf(err, "noteext dao upsert tx failed when updating note")
	}

	return newNote, nil
}

// 统一处理image和video资源处理逻辑
func (b *NoteCreatorBiz) handleNoteAssets(
	ctx context.Context,
	newNote *model.Note,
	req *CreateNoteRequest,
	videoUpdate bool, // video更新场景 用来判断是否需要更新资源
) (bool, error) {
	// 找出旧资源
	oldAssetsPO, err := b.getOldNoteAssets(ctx, newNote.NoteId, newNote.Type)
	if err != nil {
		return false, xerror.Wrapf(err, "biz get old note assets failed")
	}

	hasOldAssets := len(oldAssetsPO) > 0
	// 视频更新场景：没有传新的 fileId
	if videoUpdate && req.Video.FileId == "" && req.Video.CoverFileId == "" {
		if !hasOldAssets {
			// 没有旧资源，也没有新资源
			return false, global.ErrVideoNoteAssetNotExist
		}
		// 有旧资源，不需要更新
		return false, nil
	}

	newAssetsPO := assignNoteAssets(newNote, req)
	if len(newAssetsPO) == 0 {
		return false, nil
	}

	oldAssetMap := make(map[string]struct{}, len(oldAssetsPO))
	for _, old := range oldAssetsPO {
		oldAssetMap[old.AssetKey] = struct{}{}
	}

	// 找出old和new的资源差异
	diffAssets := make([]*notedao.AssetPO, 0, len(newAssetsPO))
	for _, asset := range newAssetsPO {
		if _, ok := oldAssetMap[asset.AssetKey]; !ok {
			diffAssets = append(diffAssets, asset)
		}
	}

	if len(diffAssets) == 0 {
		return false, nil
	}

	newAssetKeys := make([]string, 0, len(newAssetsPO))
	for _, asset := range newAssetsPO {
		newAssetKeys = append(newAssetKeys, asset.AssetKey)
	}

	if len(oldAssetsPO) > 0 {
		// 随后删除旧资源 除了newAssetKeys之外的其它
		err = b.data.NoteAsset.DeleteByNoteIdExcept(ctx, newNote.NoteId, newAssetKeys)
		if err != nil {
			return false, xerror.Wrapf(err, "noteasset dao delete tx failed")
		}
	}

	err = b.data.NoteAsset.BatchInsert(ctx, diffAssets)
	if err != nil {
		return false, xerror.Wrapf(err, "noteasset dao batch insert tx failed")
	}

	return true, nil
}

func (b *NoteCreatorBiz) getOldNoteAssets(
	ctx context.Context, noteId int64, noteType model.NoteType,
) ([]*notedao.AssetPO, error) {
	switch noteType {
	case model.AssetTypeImage:
		assets, err := b.data.NoteAsset.FindImageNoteAssets(ctx, noteId)
		if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrapf(err, "noteasset dao find image assetsfailed")
		}
		return assets, nil
	case model.AssetTypeVideo:
		assets, err := b.data.NoteAsset.FindVideoNoteAssets(ctx, noteId)
		if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrapf(err, "noteasset dao find video assets failed")
		}
		return assets, nil
	default:
		return nil, global.ErrUnsupportedResource
	}
}

func (b *NoteCreatorBiz) DeleteNote(ctx context.Context, req *DeleteNoteRequest) error {
	var (
		uid    int64 = metadata.Uid(ctx)
		noteId       = req.NoteId
	)

	queried, err := b.data.Note.FindOneForUpdate(ctx, noteId) // lock record here
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", req).WithCtx(ctx)
	}

	if uid != queried.Owner {
		return global.ErrNotNoteOwner
	}

	err = b.data.Note.Delete(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "dao delete note basic tx failed")
	}

	err = b.data.NoteAsset.DeleteByNoteId(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "dao delete note asset tx failed")
	}

	err = b.data.NoteExt.Delete(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "dao delete note ext tx failed")
	}

	return nil
}

func (b *NoteCreatorBiz) CreatorGetNote(ctx context.Context, noteId int64) (*model.Note, error) {
	var (
		uid = metadata.Uid(ctx)
		nid = noteId
	)

	note, err := b.data.Note.FindOne(ctx, nid)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz get note failed")
	}

	if uid != note.Owner {
		return nil, global.ErrNotNoteOwner
	}

	res, err := b.note.AssembleNotes(ctx, convert.NoteFromDao(note).AsSlice())
	if err != nil || len(res.Items) == 0 {
		return nil, xerror.Wrapf(err, "assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	err = b.note.AssembleNotesExt(ctx, res.Items)
	if err != nil {
		return nil, xerror.Wrapf(err, "assemble note ext failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return res.Items[0], nil
}

func (b *NoteCreatorBiz) ListNote(ctx context.Context) (*model.Notes, error) {
	uid := metadata.Uid(ctx)

	notes, err := b.data.Note.ListByOwner(ctx, uid)
	if errors.Is(err, xsql.ErrNoRecord) {
		return &model.Notes{}, nil
	}
	if err != nil {
		return nil, xerror.Wrapf(err, "biz note list by owner failed").WithCtx(ctx)
	}

	res, err := b.note.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil, xerror.Wrapf(err, "biz note assemble note failed").WithCtx(ctx)
	}

	err = b.note.AssembleNotesExt(ctx, res.Items)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz note assemble note ext failed").WithCtx(ctx)
	}

	return res, nil
}

func (b *NoteCreatorBiz) PageListNoteWithCursor(ctx context.Context, cursor int64, count int32) (*model.Notes, model.PageResult, error) {
	var (
		uid      = metadata.Uid(ctx)
		nextPage = model.PageResult{}
	)

	if cursor == 0 {
		cursor = model.MaxCursor
	}
	notes, err := b.data.Note.ListByOwnerByCursor(ctx, uid, cursor, count)
	if errors.Is(err, xsql.ErrNoRecord) {
		return &model.Notes{}, nextPage, nil
	}
	if err != nil {
		return nil, nextPage,
			xerror.Wrapf(err, "biz note list by owner with cursor failed").
				WithCtx(ctx).
				WithExtras("cursor", cursor, "count", count)
	}

	// 计算下一次请求的游标位置
	if len(notes) > 0 {
		nextPage.NextCursor = notes[len(notes)-1].Id
		if len(notes) == int(count) {
			nextPage.HasNext = true
		}
	}

	notesResp, err := b.note.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil,
			nextPage,
			xerror.Wrapf(err, "biz note failed to assemble notes when cursor page list notes").WithCtx(ctx).
				WithExtras("cursor", cursor, "count", count)
	}
	err = b.note.AssembleNotesExt(ctx, notesResp.Items)
	if err != nil {
		return nil, nextPage, xerror.Wrapf(err, "biz note assemble note ext failed").WithCtx(ctx)
	}

	return notesResp, nextPage, nil
}

// page从1开始
func (b *NoteCreatorBiz) PageListNote(ctx context.Context, page, count int32) (*model.Notes, int64, error) {
	uid := metadata.Uid(ctx)

	total, err := b.data.Note.GetPostedCountByOwner(ctx, uid)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, 0, xerror.Wrapf(err, "biz note count by owner failed").WithCtx(ctx)
		}

		return &model.Notes{}, 0, nil
	}

	notes, err := b.data.Note.PageListByOwner(ctx, uid, page, count)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "biz note page list failed").WithCtx(ctx)
	}

	notesResp, err := b.note.AssembleNotes(ctx, convert.NoteSliceFromDao(notes))
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "biz note failed to assemble notes when page list notes")
	}

	err = b.note.AssembleNotesExt(ctx, notesResp.Items)
	if err != nil {
		return nil, total, xerror.Wrapf(err, "biz note assemble note ext failed").WithCtx(ctx)
	}

	return notesResp, total, nil
}

// 新增笔记标签
func (b *NoteCreatorBiz) AddTag(ctx context.Context, name string) (int64, error) {
	id, err := b.data.Tag.Create(ctx, &tagdao.Tag{Name: name})
	if err != nil {
		if errors.Is(err, xsql.ErrDuplicate) {
			// already exist
			got, err := b.data.Tag.FindByName(ctx, name)
			if err != nil {
				return 0, xerror.Wrapf(err, "tag dao failed to find by name").WithExtra("name", name)
			}

			return got.Id, nil
		}

		return 0, xerror.Wrapf(err, "tag dao failed to create").WithExtra("name", name)
	}

	return id, nil
}

// 获取用户发布的笔记数量
func (b *NoteCreatorBiz) GetUserPostedCount(ctx context.Context, uid int64) (int64, error) {
	cnt, err := b.data.Note.GetPostedCountByOwner(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "note dao get posted count failed").
			WithExtra("uid", uid).
			WithCtx(ctx)
	}

	return cnt, nil
}

// 获取用户公开发布的笔记数量
func (b *NoteCreatorBiz) GetUserPublicPostedCount(ctx context.Context, uid int64) (int64, error) {
	cnt, err := b.data.Note.GetPublicPostedCountByOwner(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "note dao get public posted count failed").
			WithExtra("uid", uid).
			WithCtx(ctx)
	}

	return cnt, nil
}

func (b *NoteCreatorBiz) upgradeNoteState(ctx context.Context, noteId int64, state model.NoteState) error {
	err := b.data.Note.UpgradeState(ctx, noteId, state)
	if err != nil {
		return xerror.Wrapf(err, "note dao update state failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return nil
}

// 设置笔记状态为处理中
func (b *NoteCreatorBiz) TransferNoteStateToProcessing(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateProcessing)
}

// 设置笔记状态为处理完成
func (b *NoteCreatorBiz) TransferNoteStateToProcessed(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateProcessed)
}

// 设置笔记状态为处理失败
func (b *NoteCreatorBiz) TransferNoteStateToProcessFailed(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateProcessFailed)
}

// 设置笔记状态为审核中
func (b *NoteCreatorBiz) TransferNoteStateToAuditing(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateAuditing)
}

// 设置笔记状态为审核通过
func (b *NoteCreatorBiz) TransferNoteStateToAuditPassed(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateAuditPassed)
}

// 设置笔记状态为已发布
func (b *NoteCreatorBiz) TransferNoteStateToPublished(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStatePublished)
}

// 设置笔记状态为审核不通过
func (b *NoteCreatorBiz) TransferNoteStateToRejected(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateRejected)
}

// 设置笔记状态为被封禁
func (b *NoteCreatorBiz) TransferNoteStateToBanned(ctx context.Context, noteId int64) error {
	return b.upgradeNoteState(ctx, noteId, model.NoteStateBanned)
}

func (b *NoteCreatorBiz) BatchUpdateAssetMeta(ctx context.Context, noteId int64, metas map[string][]byte) error {
	err := b.data.NoteAsset.BatchUpdateAssetMeta(ctx, noteId, metas)
	if err != nil {
		return xerror.Wrapf(err, "note asset dao failed to batch update asset meta").
			WithExtra("noteId", noteId).WithCtx(ctx)
	}
	return nil
}
