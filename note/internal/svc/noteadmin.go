package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/infra/repo"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"
	noteassetrepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/noteasset"
	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type NoteAdminSvc struct {
	ctx       *ServiceContext
	repo      *repo.Repo
	OssSigner *signer.Signer
}

func NewNoteAdminSvc(ctx *ServiceContext) *NoteAdminSvc {
	return &NoteAdminSvc{
		ctx:  ctx,
		repo: infra.Repo(),
		OssSigner: signer.NewSigner(
			ctx.Config.Oss.User,
			ctx.Config.Oss.Pass,
			signer.Config{
				Endpoint: ctx.Config.Oss.DisplayEndpoint,
				Location: ctx.Config.Oss.Location,
			}),
	}
}

// 新建笔记
func (s *NoteAdminSvc) Create(ctx context.Context, req *notemodel.CreateReq) (uint64, error) {
	var (
		uid    uint64 = metadata.Uid(ctx)
		noteId uint64
	)

	now := time.Now().Unix()
	newNote := &noterepo.Model{
		Title:   req.Basic.Title,
		Desc:    req.Basic.Desc,
		Privacy: int8(req.Basic.Privacy),
		Owner:   uid,
	}

	err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 插入图片基础内容
		var err error
		noteId, err = s.repo.NoteRepo.InsertTx(ctx, tx, newNote)
		if err != nil {
			return err
		}

		// 插入笔记资源数据
		var noteAssets = make([]*noteassetrepo.Model, 0, len(req.Images))
		for _, img := range req.Images {
			noteAssets = append(noteAssets, &noteassetrepo.Model{
				AssetKey:  img.FileId,
				AssetType: global.AssetTypeImage,
				NoteId:    noteId,
				CreateAt:  now,
			})
		}
		err = s.repo.NoteAssetRepo.BatchInsertTx(ctx, tx, noteAssets)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, xerror.Wrapf(err, "repo transact insert note failed").WithExtra("req", req).WithCtx(ctx)
	}

	return noteId, nil
}

// 更新笔记
func (s *NoteAdminSvc) Update(ctx context.Context, req *notemodel.UpdateReq) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	now := time.Now().Unix()
	noteId := req.NoteId
	xlog.Msg("creator updating").Extra("noteId", noteId).Debugx(ctx)
	queried, err := s.repo.NoteRepo.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", req).WithCtx(ctx)
	}

	// 确保更新者uid和笔记作者uid相同
	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	newNote := &noterepo.Model{
		Id:       noteId,
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  int8(req.Basic.Privacy),
		Owner:    queried.Owner,
		CreateAt: queried.CreateAt,
		UpdateAt: now,
	}

	defer func() {
		// 删除缓存
		if err := CacheDelNote(ctx, noteId); err != nil {
			xlog.Msg("cache del note failed").Err(err).Extra("noteId", noteId).Errorx(ctx)
		}
	}()

	// 开启事务执行
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 先更新基础信息
		err := s.repo.NoteRepo.UpdateTx(ctx, tx, newNote)
		if err != nil {
			return xerror.Wrapf(err, "note repo update tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}

		oldAssets, err := s.repo.NoteAssetRepo.FindByNoteIdTx(ctx, tx, noteId)
		if err != nil && !errors.Is(xsql.ErrNoRecord, err) {
			return xerror.Wrapf(err, "noteasset repo find failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}
		newAssetKeys := make([]string, 0, len(req.Images))
		for _, img := range req.Images {
			newAssetKeys = append(newAssetKeys, img.FileId)
		}

		// 随后删除旧资源
		err = s.repo.NoteAssetRepo.ExcludeDeleteByNoteIdTx(ctx, tx, noteId, newAssetKeys)
		if err != nil {
			return xerror.Wrapf(err, "noteasset repo delete tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}

		// 找出old和new的资源差异，只更新发生了变化的部分
		oldAssetMap := make(map[string]struct{})
		for _, old := range oldAssets {
			oldAssetMap[old.AssetKey] = struct{}{}
		}
		newAssets := make([]*noteassetrepo.Model, 0, len(req.Images))
		for _, img := range req.Images {
			if _, ok := oldAssetMap[img.FileId]; !ok {
				newAssets = append(newAssets, &noteassetrepo.Model{
					AssetKey:  img.FileId,
					AssetType: global.AssetTypeImage,
					NoteId:    noteId,
					CreateAt:  now,
				})
			}
		}

		if len(newAssets) == 0 {
			return nil
		}

		// 插入新的资源
		err = s.repo.NoteAssetRepo.BatchInsertTx(ctx, tx, newAssets)
		if err != nil {
			return xerror.Wrapf(err, "noteasset repo batch insert tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}

		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "repo transact update note tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return nil
}

func (s *NoteAdminSvc) UploadAuth(ctx context.Context, req *notemodel.UploadAuthReq) (*notemodel.UploadAuthRes, error) {
	// 生成count个上传凭证
	fileId := s.ctx.OssKeyGen.Gen()

	now := time.Now()
	currentTime := now.Unix()

	// 生成签名
	info, err := s.OssSigner.Sign(fileId, req.MimeType)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrPermDenied.Msg("服务器签名失败"), err.Error()).WithExtra("fileId", fileId).WithCtx(ctx)
	}

	res := notemodel.UploadAuthRes{
		FildId:      fileId,
		CurrentTime: currentTime,
		ExpireTime:  info.ExpireAt.Unix(),
		UploadAddr:  s.ctx.Config.Oss.DisplayEndpoint,
		Headers: notemodel.UploadAuthResHeaders{
			Auth:   info.Auth,
			Date:   info.Date,
			Sha256: info.Sha256,
			Token:  info.Token,
		},
	}

	return &res, nil
}

// 删除笔记
func (s *NoteAdminSvc) Delete(ctx context.Context, req *notemodel.DeleteReq) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	noteId := req.NoteId
	if noteId <= 0 {
		return global.ErrNoteNotFound
	}

	queried, err := s.repo.NoteRepo.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", req).WithCtx(ctx)
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	defer func() {
		if err := CacheDelNote(ctx, noteId); err != nil {
			xlog.Msg("cache del note failed").Err(err).Extra("noteId", noteId).Errorx(ctx)
		}
	}()

	// 开始删除
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		err := s.repo.NoteRepo.DeleteTx(ctx, tx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "repo delete note basic tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}

		// err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(id, nil)(ctx, sess)
		err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(ctx, tx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "repo delete note asset tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
		}

		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "repo delete note tx failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return nil
}

// 列出某用户所有笔记
func (s *NoteAdminSvc) List(ctx context.Context) (*notemodel.BatchNoteItem, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	notes, err := s.repo.NoteRepo.ListByOwner(ctx, uid)
	if errors.Is(xsql.ErrNoRecord, err) {
		return &notemodel.BatchNoteItem{}, nil
	}
	if err != nil {
		return nil, xerror.Wrapf(err, "repo note list by owner failed").WithCtx(ctx)
	}

	return AssembleNotes(ctx, notes)
}

// 用于笔记作者获取笔记的详细信息
func (s *NoteAdminSvc) GetNote(ctx context.Context, noteId uint64) (*notemodel.Item, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
		nid        = noteId
	)

	if nid <= 0 {
		return nil, global.ErrNoteNotFound
	}

	note, err := GetNote(ctx, nid)
	if err != nil {
		return nil, xerror.Wrapf(err, "GetNote failed")
	}

	if uid != note.Owner {
		return nil, global.ErrNotNoteOwner
	}

	res, err := AssembleNotes(ctx, []*noterepo.Model{note})
	if err != nil || len(res.Items) == 0 {
		return nil, xerror.Wrapf(err, "assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return res.Items[0], nil
}

// 获取笔记作者
func (s *NoteAdminSvc) GetNoteOwner(ctx context.Context, nid uint64) (uint64, error) {
	if nid <= 0 {
		return 0, global.ErrNoteNotFound
	}

	n, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return 0, xerror.Wrapf(err, "note repo find one failed").WithExtra("noteId", nid).WithCtx(ctx)
		}
		return 0, global.ErrNoteNotFound
	}

	return n.Owner, nil
}
