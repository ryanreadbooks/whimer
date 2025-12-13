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

	var noteAssets []*notedao.AssetPO
	switch req.Basic.NoteType {
	case model.AssetTypeImage:
		noteAssets = make([]*notedao.AssetPO, 0, len(req.Images))
		for _, img := range req.Images {
			imgMeta := model.NewAssetImageMeta(img.Width, img.Height, img.Format).Bytes()
			noteAssets = append(noteAssets, &notedao.AssetPO{
				AssetKey:  img.FileId,           // 包含桶名称
				AssetType: model.AssetTypeImage, // image
				CreateAt:  now,
				AssetMeta: imgMeta,
			})
		}
	case model.AssetTypeVideo:
		// TODO videos
	}

	var ext = notedao.ExtPO{
		AtUsers: json.RawMessage{},
	}

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
		for _, a := range noteAssets {
			a.NoteId = noteId
		}

		// 插入笔记资源数据
		errTx = b.data.NoteAsset.BatchInsert(ctx, noteAssets)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note asset dao batch insert tx failed")
		}

		ext.NoteId = noteId

		// 笔记额外信息
		if len(req.TagIds) > 0 {
			tagIdList := xslice.JoinInts(req.TagIds)
			ext.Tags = tagIdList
		}
		if len(req.AtUsers) > 0 {
			if data, err := json.Marshal(req.AtUsers); err == nil {
				ext.AtUsers = data
			}
		}

		if isNoteExtValid(&ext) {
			errTx = b.data.NoteExt.Upsert(ctx, &ext)
			if errTx != nil {
				return xerror.Wrapf(errTx, "note ext insert tx failed")
			}
		}

		return nil
	})

	if err != nil {
		return nil, xerror.Wrapf(err, "biz create note failed").WithExtra("note", req).WithCtx(ctx)
	}

	newNote := convert.NoteFromDao(newNotePO)

	return newNote, nil
}

func (b *NoteCreatorBiz) UpdateNote(ctx context.Context, req *UpdateNoteRequest) error {
	var (
		uid = metadata.Uid(ctx)
		ip  = xnet.IpAsBytes(metadata.ClientIp(ctx))
	)

	now := time.Now().Unix()
	noteId := req.NoteId
	oldNote, err := b.data.Note.FindOneForUpdate(ctx, noteId)
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "biz find one note failed").WithExtra("note", req).WithCtx(ctx)
	}

	if oldNote.NoteType != req.Basic.NoteType {
		return global.ErrNoteTypeCannotChange
	}

	// 确保更新者uid和笔记作者uid相同
	if uid != oldNote.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	newNote := &notedao.NotePO{
		Id:       noteId,
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  req.Basic.Privacy,
		NoteType: oldNote.NoteType,
		State:    model.NoteStateInit, // 更新后需要重新走流程
		CreateAt: oldNote.CreateAt,
		UpdateAt: now,
		Owner:    oldNote.Owner,
		Ip:       ip,
	}

	var newAssetPos []*notedao.AssetPO
	switch req.Basic.NoteType {
	case model.AssetTypeImage:
		newAssetPos = make([]*notedao.AssetPO, 0, len(req.Images))
		for _, img := range req.Images {
			imgMeta := model.NewAssetImageMeta(img.Width, img.Height, img.Format).Bytes()
			newAssetPos = append(newAssetPos, &notedao.AssetPO{
				AssetKey:  img.FileId,
				AssetType: model.AssetTypeImage,
				NoteId:    noteId,
				CreateAt:  now,
				AssetMeta: imgMeta,
			})
		}
	case model.AssetTypeVideo:
		// TODO
	}

	// 先更新基础信息
	err = b.data.Note.Update(ctx, newNote)
	if err != nil {
		return xerror.Wrapf(err, "note dao update tx failed")
	}

	// 找出旧资源
	oldAssets, err := b.data.NoteAsset.FindImageByNoteId(ctx, newNote.Id)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		return xerror.Wrapf(err, "noteasset dao find failed")
	}

	// 笔记的新资源
	newAssetKeys := make([]string, 0, len(newAssetPos))
	for _, asset := range newAssetPos {
		newAssetKeys = append(newAssetKeys, asset.AssetKey)
	}

	// 随后删除旧资源
	// 删除除了newAssetKeys之外的其它
	err = b.data.NoteAsset.ExcludeDeleteImageByNoteId(ctx, newNote.Id, newAssetKeys)
	if err != nil {
		return xerror.Wrapf(err, "noteasset dao delete tx failed")
	}

	// 找出old和new的资源差异，只更新发生了变化的部分
	oldAssetMap := make(map[string]struct{})
	for _, old := range oldAssets {
		oldAssetMap[old.AssetKey] = struct{}{}
	}
	newAssets := make([]*notedao.AssetPO, 0, len(newAssetPos))
	for _, asset := range newAssetPos {
		if _, ok := oldAssetMap[asset.AssetKey]; !ok {
			newAssets = append(newAssets, &notedao.AssetPO{
				AssetKey:  asset.AssetKey,
				AssetType: model.AssetTypeImage,
				NoteId:    newNote.Id,
				CreateAt:  now,
				AssetMeta: asset.AssetMeta,
			})
		}
	}

	if len(newAssets) == 0 {
		return nil
	}

	// 插入新的资源
	err = b.data.NoteAsset.BatchInsert(ctx, newAssets)
	if err != nil {
		return xerror.Wrapf(err, "noteasset dao batch insert tx failed")
	}

	// ext处理
	ext := notedao.ExtPO{
		NoteId:  oldNote.Id,
		Tags:    xslice.JoinInts(req.TagIds),
		AtUsers: json.RawMessage{},
	}
	if len(req.AtUsers) > 0 {
		if data, err := json.Marshal(req.AtUsers); err == nil {
			ext.AtUsers = data
		}
	}

	if isNoteExtValid(&ext) {
		err = b.data.NoteExt.Upsert(ctx, &ext)
		if err != nil {
			return xerror.Wrapf(err, "noteext dao upsert tx failed when updating note")
		}
	}

	return nil
}

func (b *NoteCreatorBiz) DeleteNote(ctx context.Context, req *DeleteNoteRequest) error {
	var (
		uid    int64 = metadata.Uid(ctx)
		noteId       = req.NoteId
	)

	queried, err := b.data.Note.FindOne(ctx, noteId)
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", req).WithCtx(ctx)
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
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
	var (
		uid = metadata.Uid(ctx)
	)

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
	var (
		uid = metadata.Uid(ctx)
	)

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

func (b *NoteCreatorBiz) setNoteState(ctx context.Context, noteId int64, state model.NoteState) error {
	err := b.data.Note.UpdateState(ctx, noteId, state)
	if err != nil {
		return xerror.Wrapf(err, "note dao update state failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return nil
}

// 设置笔记状态为处理中
func (b *NoteCreatorBiz) SetNoteStateProcessing(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateProcessing)
}

// 设置笔记状态为处理完成
func (b *NoteCreatorBiz) SetNoteStateProcessed(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateProcessed)
}

// 设置笔记状态为处理失败
func (b *NoteCreatorBiz) SetNoteStateProcessFailed(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateProcessFailed)
}

// 设置笔记状态为审核中
func (b *NoteCreatorBiz) SetNoteStateAuditing(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateAuditing)
}

// 设置笔记状态为审核通过
func (b *NoteCreatorBiz) SetNoteStateAuditPassed(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateAuditPassed)
}

// 设置笔记状态为已发布
func (b *NoteCreatorBiz) SetNoteStatePublished(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStatePublished)
}

// 设置笔记状态为审核不通过
func (b *NoteCreatorBiz) SetNoteStateRejected(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateRejected)
}

// 设置笔记状态为被封禁
func (b *NoteCreatorBiz) SetNoteStateBanned(ctx context.Context, noteId int64) error {
	return b.setNoteState(ctx, noteId, model.NoteStateBanned)
}
