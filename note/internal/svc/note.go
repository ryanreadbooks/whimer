package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	mnote "github.com/ryanreadbooks/whimer/misc/note"
	"github.com/ryanreadbooks/whimer/misc/oss"
	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	crtp "github.com/ryanreadbooks/whimer/note/internal/model/note"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/repo/note"
	noteassetrepo "github.com/ryanreadbooks/whimer/note/internal/repo/noteasset"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type NoteSvc struct {
	repo *repo.Repo

	Ctx            *ServiceContext
	NoteIdConfuser *safety.Confuser
	OssSigner      *signer.Signer
}

func NewNoteSvc(ctx *ServiceContext, repo *repo.Repo) *NoteSvc {
	return &NoteSvc{
		repo:           repo,
		Ctx:            ctx,
		NoteIdConfuser: mnote.NewConfuser(),
		OssSigner: signer.NewSigner(
			ctx.Config.Oss.User,
			ctx.Config.Oss.Pass,
			signer.Config{
				Endpoint: ctx.Config.Oss.DisplayEndpoint,
				Location: ctx.Config.Oss.Location,
			}),
	}
}

func (s *NoteSvc) Create(ctx context.Context, req *crtp.CreateReq) (uint64, error) {
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
		logx.Errorf("repo transact insert note err: %v, req: %+v, uid: %d", err, req, uid)
		return 0, global.ErrInsertNoteFail
	}

	return noteId, nil
}

func (s *NoteSvc) Update(ctx context.Context, req *crtp.UpdateReq) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	now := time.Now().Unix()
	noteId := req.NoteId
	logx.Debugf("creator updating noteid: %d", noteId)
	queried, err := s.repo.NoteRepo.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		logx.Errorf("repo find one note err: %v, req: %+v, uid: %d", err, req, uid)
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

	// 开启事务执行
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		// 先更新基础信息
		err := s.repo.NoteRepo.UpdateTx(ctx, tx, newNote)
		if err != nil {
			logx.Errorf("note repo update tx err: %v, noteid: %d", err, noteId)
			return err
		}

		oldAssets, err := s.repo.NoteAssetRepo.FindByNoteIdTx(ctx, tx, noteId)
		if err != nil && !errors.Is(xsql.ErrNoRecord, err) {
			logx.Errorf("noteasset repo find err: %v, noteid: %d", err, noteId)
			return err
		}
		newAssetKeys := make([]string, 0, len(req.Images))
		for _, img := range req.Images {
			newAssetKeys = append(newAssetKeys, img.FileId)
		}

		// 随后删除旧资源
		err = s.repo.NoteAssetRepo.ExcludeDeleteByNoteIdTx(ctx, tx, noteId, newAssetKeys)
		if err != nil {
			logx.Errorf("noteasset repo delete tx err: %v, noteid: %d", err, noteId)
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
			logx.Errorf("noteasset repo batch insert tx err: %v, noteid: %d", err, noteId)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("repo transact update note err: %v, id: %d", err, noteId)
		return global.ErrUpdateNoteFail
	}

	return nil
}

func (s *NoteSvc) UploadAuth(ctx context.Context, req *crtp.UploadAuthReq) (*crtp.UploadAuthRes, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	// 生成count个上传凭证
	fileId := s.Ctx.OssKeyGen.Gen()

	now := time.Now()
	currentTime := now.Unix()

	// 生成签名
	info, err := s.OssSigner.Sign(fileId, req.MimeType)
	if err != nil {
		logx.Errorf("upload auth sign err: %v, fileid: %s, uid: %d", err, fileId, uid)
		return nil, global.ErrPermDenied.Msg("服务器签名失败")
	}

	res := crtp.UploadAuthRes{
		FildId:      fileId,
		CurrentTime: currentTime,
		ExpireTime:  info.ExpireAt.Unix(),
		UploadAddr:  s.Ctx.Config.Oss.DisplayEndpoint,
		Headers: crtp.UploadAuthResHeaders{
			Auth:   info.Auth,
			Date:   info.Date,
			Sha256: info.Sha256,
			Token:  info.Token,
		},
	}

	return &res, nil
}

func (s *NoteSvc) Delete(ctx context.Context, req *crtp.DeleteReq) error {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	noteId := req.NoteId
	if noteId <= 0 {
		return global.ErrNoteNotFound
	}

	logx.Debugf("creator deleting noteid: %d", noteId)
	queried, err := s.repo.NoteRepo.FindOne(ctx, noteId)
	if errors.Is(xsql.ErrNoRecord, err) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		logx.Errorf("repo find one note err: %v, req: %+v, uid: %d", err, req, uid)
		return global.ErrDeleteNoteFail
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	// 开始删除
	err = s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
		err := s.repo.NoteRepo.DeleteTx(ctx, tx, noteId)
		if err != nil {
			logx.Errorf("repo delete note basic tx err: %v, noteid: %d", err, noteId)
			return err
		}

		// err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(id, nil)(ctx, sess)
		err = s.repo.NoteAssetRepo.DeleteByNoteIdTx(ctx, tx, noteId)
		if err != nil {
			logx.Errorf("repo delete note asset tx err: %v, noteid: %d", err, noteId)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("repo delete note tx err: %v, noteid: %d", err, noteId)
		return global.ErrDeleteNoteFail
	}

	return nil
}

func (s *NoteSvc) List(ctx context.Context) (*crtp.ListRes, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
	)

	notes, err := s.repo.NoteRepo.ListByOwner(ctx, uid)
	if errors.Is(xsql.ErrNoRecord, err) {
		return &crtp.ListRes{}, nil
	}

	if err != nil {
		logx.Errorf("repo note list by owner err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	var noteIds = make([]uint64, 0, len(notes))
	for _, note := range notes {
		noteIds = append(noteIds, note.Id)
	}

	// 获取资源信息
	noteAssets, err := s.repo.NoteAssetRepo.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		logx.Errorf("repo note list by owner err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	// 组合notes和noteAssets
	var res crtp.ListRes
	for _, note := range notes {
		item := &crtp.ListResItem{
			NoteId:   note.Id,
			Title:    note.Title,
			Desc:     note.Desc,
			Privacy:  note.Privacy,
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		for _, asset := range noteAssets {
			if note.Id == asset.NoteId {
				item.Images = append(item.Images, &crtp.ListResItemImage{
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

	return &res, nil
}

func (s *NoteSvc) GetNote(ctx context.Context, noteId uint64) (*crtp.ListResItem, error) {
	var (
		uid uint64 = metadata.Uid(ctx)
		nid        = noteId
	)

	if nid <= 0 {
		return nil, global.ErrNoteNotFound
	}

	note, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if errors.Is(err, xsql.ErrNoRecord) {
		return nil, global.ErrNoteNotFound
	}

	if err != nil {
		logx.Errorf("repo note find one err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	if note.Owner != uid {
		return nil, global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	assets, err := s.repo.NoteAssetRepo.FindByNoteIds(ctx, []uint64{note.Id})
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		logx.Errorf("repo note asset find by note ids err: %v, noteid: %d", err, note.Id)
		return nil, global.ErrGetNoteFail
	}

	var res = crtp.ListResItem{
		NoteId:   noteId,
		Title:    note.Title,
		Desc:     note.Desc,
		Privacy:  note.Privacy,
		CreateAt: note.CreateAt,
		UpdateAt: note.UpdateAt,
	}

	for _, asset := range assets {
		res.Images = append(res.Images, &crtp.ListResItemImage{
			Url: oss.GetPublicVisitUrl(
				s.Ctx.Config.Oss.Bucket,
				asset.AssetKey,
				s.Ctx.Config.Oss.DisplayEndpoint,
			),
			Type: int(asset.AssetType),
		})
	}

	return &res, nil
}

func (s *NoteSvc) IsNoteExist(ctx context.Context, noteId string) (bool, error) {
	nid := s.NoteIdConfuser.DeConfuseU(noteId)
	if nid <= 0 {
		return false, global.ErrNoteNotFound
	}

	_, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			logx.Errorf("note repo find one err: %v, nid: %d", err, nid)
			return false, global.ErrInternal
		}
		return false, global.ErrNoteNotFound
	}

	return true, nil
}

func (s *NoteSvc) GetNoteOwner(ctx context.Context, noteId string) (uint64, error) {
	nid := s.NoteIdConfuser.DeConfuseU(noteId)
	if nid <= 0 {
		return 0, global.ErrNoteNotFound
	}

	n, err := s.repo.NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			logx.Errorf("note repo find one err: %v, nid: %d", err, nid)
			return 0, global.ErrInternal
		}
		return 0, global.ErrNoteNotFound
	}

	return n.Owner, nil
}
