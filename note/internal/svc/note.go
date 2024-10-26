package svc

import (
	"context"
	"errors"
	"time"

	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concur"
	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/oss"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/external"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/repo/note"
	noteassetrepo "github.com/ryanreadbooks/whimer/note/internal/repo/noteasset"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type NoteSvc struct {
	repo           *repo.Repo
	cache          *NoteCache
	Ctx            *ServiceContext
	NoteIdConfuser *safety.Confuser
	OssSigner      *signer.Signer
}

func NewNoteSvc(ctx *ServiceContext, repo *repo.Repo, cache *redis.Redis) *NoteSvc {
	return &NoteSvc{
		repo:           repo,
		cache:          NewNoteCache(cache),
		Ctx:            ctx,
		NoteIdConfuser: safety.NewConfuser(ctx.Config.Salt, 24), // TODO can be removed
		OssSigner: signer.NewSigner(
			ctx.Config.Oss.User,
			ctx.Config.Oss.Pass,
			signer.Config{
				Endpoint: ctx.Config.Oss.DisplayEndpoint,
				Location: ctx.Config.Oss.Location,
			}),
	}
}

func (s *NoteSvc) Create(ctx context.Context, req *notemodel.CreateReq) (uint64, error) {
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
		xlog.Msg("repo transact insert note err").Err(err).Extra("req", req).Errorx(ctx)
		return 0, global.ErrInsertNoteFail
	}

	return noteId, nil
}

func (s *NoteSvc) Update(ctx context.Context, req *notemodel.UpdateReq) error {
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
		xlog.Msg("repo find one note err").Err(err).Extra("req", req).Errorx(ctx)
		return global.ErrUpdateNoteFail
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
		if err := s.cache.DelNote(ctx, noteId); err != nil {
			xlog.Msg("cache del note failed").Err(err).Extra("noteId", noteId).Errorx(ctx)
		}
	}()

	// 开启事务执行
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 先更新基础信息
		err := s.repo.NoteRepo.UpdateTx(ctx, tx, newNote)
		if err != nil {
			xlog.Msg("note repo update tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
		}

		oldAssets, err := s.repo.NoteAssetRepo.FindByNoteIdTx(ctx, tx, noteId)
		if err != nil && !errors.Is(xsql.ErrNoRecord, err) {
			xlog.Msg("noteasset repo find err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
		}
		newAssetKeys := make([]string, 0, len(req.Images))
		for _, img := range req.Images {
			newAssetKeys = append(newAssetKeys, img.FileId)
		}

		// 随后删除旧资源
		err = s.repo.NoteAssetRepo.ExcludeDeleteByNoteIdTx(ctx, tx, noteId, newAssetKeys)
		if err != nil {
			xlog.Msg("noteasset repo delete tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
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
			xlog.Msg("noteasset repo batch insert tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
		}

		return nil
	})

	if err != nil {
		xlog.Msg("repo transact update note tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
		return global.ErrUpdateNoteFail
	}

	return nil
}

func (s *NoteSvc) UploadAuth(ctx context.Context, req *notemodel.UploadAuthReq) (*notemodel.UploadAuthRes, error) {
	// 生成count个上传凭证
	fileId := s.Ctx.OssKeyGen.Gen()

	now := time.Now()
	currentTime := now.Unix()

	// 生成签名
	info, err := s.OssSigner.Sign(fileId, req.MimeType)
	if err != nil {
		xlog.Msg("upload auth sign err").Err(err).Extra("fileId", fileId).Errorx(ctx)
		return nil, global.ErrPermDenied.Msg("服务器签名失败")
	}

	res := notemodel.UploadAuthRes{
		FildId:      fileId,
		CurrentTime: currentTime,
		ExpireTime:  info.ExpireAt.Unix(),
		UploadAddr:  s.Ctx.Config.Oss.DisplayEndpoint,
		Headers: notemodel.UploadAuthResHeaders{
			Auth:   info.Auth,
			Date:   info.Date,
			Sha256: info.Sha256,
			Token:  info.Token,
		},
	}

	return &res, nil
}

func (s *NoteSvc) Delete(ctx context.Context, req *notemodel.DeleteReq) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	noteId := req.NoteId
	if noteId <= 0 {
		return global.ErrNoteNotFound
	}

	xlog.Msg("creator deleting").Extra("noteId", noteId).Debugx(ctx)
	queried, err := s.repo.NoteRepo.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		xlog.Msg("repo find one note err").Err(err).Extra("req", req).Errorx(ctx)
		return global.ErrDeleteNoteFail
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	defer func() {
		if err := s.cache.DelNote(ctx, noteId); err != nil {
			xlog.Msg("cache del note failed").Err(err).Extra("noteId", noteId).Errorx(ctx)
		}
	}()

	// 开始删除
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		err := s.repo.NoteRepo.DeleteTx(ctx, tx, noteId)
		if err != nil {
			xlog.Msg("repo delete note basic tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
		}

		// err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(id, nil)(ctx, sess)
		err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(ctx, tx, noteId)
		if err != nil {
			xlog.Msg("repo delete note asset tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
			return err
		}

		return nil
	})

	if err != nil {
		xlog.Msg("repo delete note tx err").Err(err).Extra("noteId", noteId).Errorx(ctx)
		return global.ErrDeleteNoteFail
	}

	return nil
}

func (s *NoteSvc) List(ctx context.Context) (*notemodel.ListRes, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	notes, err := s.repo.NoteRepo.ListByOwner(ctx, uid)
	if errors.Is(xsql.ErrNoRecord, err) {
		return &notemodel.ListRes{}, nil
	}

	if err != nil {
		xlog.Msg("repo note list by owner err").Err(err).Errorx(ctx)
		return nil, global.ErrGetNoteFail
	}

	var noteIds = make([]uint64, 0, len(notes))
	likesReq := make([]*counterv1.GetSummaryRequest, 0, len(notes))
	for _, note := range notes {
		noteIds = append(noteIds, note.Id)
		likesReq = append(likesReq, &counterv1.GetSummaryRequest{
			BizCode: global.NoteLikeBizcode,
			Oid:     note.Id,
		})
	}

	// 获取资源信息
	noteAssets, err := s.repo.NoteAssetRepo.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		xlog.Msg("repo note asset list by owner err").Err(err).Errorx(ctx)
		return nil, global.ErrGetNoteFail
	}

	// 组合notes和noteAssets
	var res notemodel.ListRes
	for _, note := range notes {
		item := &notemodel.ListResItem{
			NoteId:   note.Id,
			Title:    note.Title,
			Desc:     note.Desc,
			Privacy:  note.Privacy,
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		for _, asset := range noteAssets {
			if note.Id == asset.NoteId {
				item.Images = append(item.Images, &notemodel.ListResItemImage{
					Url: oss.GetPublicVisitUrl(
						s.Ctx.Config.Oss.Bucket,
						asset.AssetKey,
						s.Ctx.Config.Oss.DisplayEndpoint,
					),
					Type: int(asset.AssetType),
				})
			}
		}

		res.Items = append(res.Items, item)
	}

	// 获取点赞数量
	likesResp, err := external.GetCounter().BatchGetSummary(ctx, &counterv1.BatchGetSummaryRequest{
		Requests: likesReq,
	})
	if err != nil {
		xlog.Msg("counter failed to batch get summary").
			Err(err).
			Extra("note_ids", noteIds).
			Infox(ctx)
	}
	if likesResp != nil {
		m := make(map[uint64]uint64, len(likesResp.Responses))
		for _, r := range likesResp.Responses {
			m[r.Oid] = r.Count
		}
		for _, item := range res.Items {
			if likeCnt, ok := m[item.NoteId]; ok {
				item.Likes = likeCnt
			}
		}
	}

	return &res, nil
}

func (s *NoteSvc) GetNote(ctx context.Context, noteId uint64) (*notemodel.ListResItem, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
		nid        = noteId
	)

	if nid <= 0 {
		return nil, global.ErrNoteNotFound
	}

	note, err := s.cache.GetNote(ctx, nid)
	if err != nil {
		xlog.Msg("cache get note failed").Err(err).Extra("noteId", nid).Errorx(ctx)
		note, err = s.repo.NoteRepo.FindOne(ctx, nid)
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrNoteNotFound
		}

		if err != nil {
			xlog.Msg("repo note find one err").Err(err).Errorx(ctx)
			return nil, global.ErrGetNoteFail
		}

		concur.SafeGo(func() {
			ctxc := context.WithoutCancel(ctx)
			if errg := s.cache.SetNote(ctxc, note); errg != nil {
				xlog.Msg("cache set note failed").Err(err).Extra("note", note).Errorx(ctxc)
			}
		})
	}

	if note.Owner != uid {
		return nil, global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	assets, err := s.repo.NoteAssetRepo.FindByNoteIds(ctx, []uint64{note.Id})
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		xlog.Msg("repo note asset find by note ids err").Err(err).Extra("noteId", note.Id).Errorx(ctx)
		return nil, global.ErrGetNoteFail
	}

	var res = notemodel.ListResItem{
		NoteId:   noteId,
		Title:    note.Title,
		Desc:     note.Desc,
		Privacy:  note.Privacy,
		CreateAt: note.CreateAt,
		UpdateAt: note.UpdateAt,
	}

	for _, asset := range assets {
		res.Images = append(res.Images, &notemodel.ListResItemImage{
			Url: oss.GetPublicVisitUrl(
				s.Ctx.Config.Oss.Bucket,
				asset.AssetKey,
				s.Ctx.Config.Oss.DisplayEndpoint,
			),
			Type: int(asset.AssetType),
		})
	}

	// 获取点赞数
	likes, err := s.GetNoteLikes(ctx, noteId)
	if err != nil {
		xlog.Msg("failed to get note likes count").Err(err).Extra("note_id", nid).Infox(ctx)
	}
	res.Likes = likes

	return &res, nil
}

func (s *NoteSvc) IsNoteExist(ctx context.Context, nid uint64) (bool, error) {
	if nid <= 0 {
		return false, global.ErrNoteNotFound
	}

	_, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			xlog.Msg("note repo find one err").Err(err).Extra("noteId", nid).Errorx(ctx)
			return false, global.ErrInternal
		}
		return false, global.ErrNoteNotFound
	}

	return true, nil
}

func (s *NoteSvc) GetNoteOwner(ctx context.Context, nid uint64) (uint64, error) {
	if nid <= 0 {
		return 0, global.ErrNoteNotFound
	}

	n, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			xlog.Msg("note repo find one err").Err(err).Extra("noteId", nid).Errorx(ctx)
			return 0, global.ErrInternal
		}
		return 0, global.ErrNoteNotFound
	}

	return n.Owner, nil
}

// 点赞笔记
func (s *NoteSvc) LikeNote(ctx context.Context, in *notev1.LikeNoteReq) (*notev1.LikeNoteRes, error) {
	var (
		opUid = metadata.Uid(ctx)
		err   error
	)

	if opUid != in.Uid {
		return nil, errorx.ErrPermission
	}

	if ok, err := s.IsNoteExist(ctx, in.NoteId); err != nil || !ok {
		return nil, err
	}

	if in.Operation == notev1.LikeNoteReq_OPERATION_UNDO_LIKE {
		// 取消点赞
		_, err = external.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     in.Uid,
			Oid:     in.NoteId,
		})
	} else {
		// 点赞
		_, err = external.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: global.NoteLikeBizcode,
			Uid:     in.Uid,
			Oid:     in.NoteId,
		})
	}

	if err != nil {
		xlog.Msg("call counter returned err").Err(err).
			Extra("op", in.Operation).
			Extra("uid", in.Uid).
			Extra("note_id", in.NoteId).
			Errorx(ctx)
		return nil, err
	}

	return &notev1.LikeNoteRes{}, nil
}

// 获取笔记点赞数量
func (s *NoteSvc) GetNoteLikes(ctx context.Context, nid uint64) (uint64, error) {
	if ok, err := s.IsNoteExist(ctx, nid); err != nil || !ok {
		return 0, err
	}

	resp, err := external.GetCounter().GetSummary(ctx, &counterv1.GetSummaryRequest{
		BizCode: global.NoteLikeBizcode,
		Oid:     nid,
	})
	if err != nil {
		xlog.Msg("counter get summary failed").
			Err(err).
			Extra("note_id", nid).
			Errorx(ctx)
		return 0, global.ErrGetNoteLikesFail
	}

	return resp.Count, nil
}
