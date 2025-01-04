package biz

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// 笔记相关
type NoteCreatorBiz interface {
	// 创作者相关
	CreatorCreateNote(ctx context.Context, note *model.CreateNoteRequest) (uint64, error)
	CreatorUpdateNote(ctx context.Context, note *model.UpdateNoteRequest) error
	CreatorDeleteNote(ctx context.Context, note *model.DeleteNoteRequest) error
	CreatorGetNote(ctx context.Context, noteId uint64) (*model.Note, error)
	CreatorListNote(ctx context.Context) (*model.Notes, error)
	CreatorPageListNote(ctx context.Context, cursor uint64, count int32) (*model.Notes, model.PageResult, error)
	CreatorGetUploadAuth(ctx context.Context, req *model.UploadAuthRequest) (*model.UploadAuthResponse, error)
}

type noteCreatorBiz struct {
	noteBiz
	OssKeyGen *keygen.Generator
	OssSigner *signer.Signer
}

func NewNoteCreatorBiz() NoteCreatorBiz {
	b := &noteCreatorBiz{
		OssKeyGen: keygen.NewGenerator(
			keygen.WithBucket(config.Conf.Oss.Bucket),
			keygen.WithPrefix(config.Conf.Oss.Prefix),
			keygen.WithPrependBucket(true),
		),
		OssSigner: signer.NewSigner(
			config.Conf.Oss.User,
			config.Conf.Oss.Pass,
			signer.Config{
				Endpoint: config.Conf.Oss.DisplayEndpoint,
				Location: config.Conf.Oss.Location,
			}),
	}

	return b
}

func (b *noteCreatorBiz) CreatorCreateNote(ctx context.Context, note *model.CreateNoteRequest) (uint64, error) {
	var (
		uid    uint64 = metadata.Uid(ctx)
		noteId uint64
	)

	now := time.Now().Unix()
	newNote := &dao.Note{
		Title:   note.Basic.Title,
		Desc:    note.Basic.Desc,
		Privacy: int8(note.Basic.Privacy),
		Owner:   uid,
	}

	var noteAssets = make([]*dao.NoteAsset, 0, len(note.Images))
	for _, img := range note.Images {
		imgMeta := model.NewAssetImageMeta(img.Width, img.Height, img.Format).String()
		noteAssets = append(noteAssets, &dao.NoteAsset{
			AssetKey:  strings.TrimLeft(img.FileId, config.Conf.Oss.Bucket+"/"), // 存储时不需要桶前缀
			AssetType: global.AssetTypeImage,
			NoteId:    noteId,
			CreateAt:  now,
			AssetMeta: imgMeta,
		})
	}

	noteId, err := infra.Dao().CreateNote(ctx, newNote, noteAssets)

	if err != nil {
		return 0, xerror.Wrapf(err, "biz create note failed").WithExtra("note", note).WithCtx(ctx)
	}

	return noteId, nil
}

func (b *noteCreatorBiz) CreatorUpdateNote(ctx context.Context, note *model.UpdateNoteRequest) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	now := time.Now().Unix()
	noteId := note.NoteId
	queried, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "biz find one note failed").WithExtra("note", note).WithCtx(ctx)
	}

	// 确保更新者uid和笔记作者uid相同
	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	newNote := &dao.Note{
		Id:       noteId,
		Title:    note.Basic.Title,
		Desc:     note.Basic.Desc,
		Privacy:  int8(note.Basic.Privacy),
		Owner:    queried.Owner,
		CreateAt: queried.CreateAt,
		UpdateAt: now,
	}

	assets := make([]*dao.NoteAsset, 0, len(note.Images))
	for _, img := range note.Images {
		assets = append(assets, &dao.NoteAsset{
			AssetKey:  img.FileId,
			AssetType: global.AssetTypeImage,
			NoteId:    noteId,
			CreateAt:  now,
		})
	}

	err = infra.Dao().UpdateNote(ctx, newNote, assets)
	if err != nil {
		return xerror.Wrapf(err, "biz update note failed").WithExtras("req", note).WithCtx(ctx)
	}

	return nil
}

func (b *noteCreatorBiz) CreatorDeleteNote(ctx context.Context, note *model.DeleteNoteRequest) error {
	var (
		uid    uint64 = metadata.Uid(ctx)
		noteId        = note.NoteId
	)

	queried, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", note).WithCtx(ctx)
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	err = infra.Dao().DeleteNote(ctx, note.NoteId)
	if err != nil {
		return xerror.Wrapf(err, "biz delete note failed").WithExtras("req", note).WithCtx(ctx)
	}

	return nil
}

func (b *noteCreatorBiz) CreatorGetNote(ctx context.Context, noteId uint64) (*model.Note, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
		nid        = noteId
	)

	note, err := infra.Dao().NoteDao.FindOne(ctx, nid)
	if err != nil {
		if xsql.IsNotFound(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz get note failed")
	}

	if uid != note.Owner {
		return nil, global.ErrNotNoteOwner
	}

	res, err := b.AssembleNotes(ctx, model.NoteFromDao(note).AsSlice())
	if err != nil || len(res.Items) == 0 {
		return nil, xerror.Wrapf(err, "assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return res.Items[0], nil
}

func (b *noteCreatorBiz) CreatorListNote(ctx context.Context) (*model.Notes, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	notes, err := infra.Dao().NoteDao.ListByOwner(ctx, uid)
	if errors.Is(xsql.ErrNoRecord, err) {
		return &model.Notes{}, nil
	}
	if err != nil {
		return nil, xerror.Wrapf(err, "biz note list by owner failed").WithCtx(ctx)
	}

	return b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
}

func (b *noteCreatorBiz) CreatorPageListNote(ctx context.Context, cursor uint64, count int32) (*model.Notes, model.PageResult, error) {
	var (
		uid      uint64 = metadata.Uid(ctx)
		nextPage        = model.PageResult{}
	)

	if cursor == 0 {
		cursor = math.MaxUint64
	}
	notes, err := infra.Dao().NoteDao.ListByOwnerByCursor(ctx, uid, cursor, count)
	if errors.Is(xsql.ErrNoRecord, err) {
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

	notesResp, err := b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
	if err != nil {
		return nil, nextPage,
			xerror.Wrapf(err, "biz note failed to assemble notes when page list notes").
				WithCtx(ctx).
				WithExtras("cursor", cursor, "count", count)
	}

	return notesResp, nextPage, nil
}

func (b *noteCreatorBiz) CreatorGetUploadAuth(ctx context.Context, req *model.UploadAuthRequest) (*model.UploadAuthResponse, error) {
	// 生成count个上传凭证
	fileId := b.OssKeyGen.Gen()

	now := time.Now()
	currentTime := now.Unix()

	// 生成签名
	info, err := b.OssSigner.Sign(fileId)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrPermDenied.Msg("服务器签名失败"), "%s", err.Error()).WithExtra("fileId", fileId).WithCtx(ctx)
	}

	res := model.UploadAuthResponse{
		FildId:      fileId,
		CurrentTime: currentTime,
		ExpireTime:  info.ExpireAt.Unix(),
		UploadAddr:  config.Conf.Oss.DisplayEndpoint,
		Headers: model.UploadAuthResponseHeaders{
			Auth:   info.Auth,
			Date:   info.Date,
			Sha256: info.Sha256,
			Token:  info.Token,
		},
	}

	return &res, nil
}
