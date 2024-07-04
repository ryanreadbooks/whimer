package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/oss/signer"
	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
	reponote "github.com/ryanreadbooks/whimer/note/internal/repo/note"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/types/manage"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	noteIdConfuserSalt = "0x7c00:noteIdConfuser:.$35%io"
)

type Manage struct {
	dao *repo.Dao

	Ctx            *ServiceContext
	NoteIdConfuser *safety.Confuser
	Signer         *signer.Signer
}

func NewManage(ctx *ServiceContext, repo *repo.Dao) *Manage {
	return &Manage{
		dao:            repo,
		Ctx:            ctx,
		NoteIdConfuser: safety.NewConfuser(noteIdConfuserSalt, 24),
		Signer: signer.NewSigner(
			ctx.Config.Oss.User,
			ctx.Config.Oss.Pass,
			signer.Config{
				Endpoint: ctx.Config.Oss.Endpoint,
				Location: ctx.Config.Oss.Location,
			}),
	}
}

func (s *Manage) Get(ctx context.Context, uid int64, noteId string) error {

	return nil
}

func (s *Manage) Create(ctx context.Context, uid int64, req *mgtp.CreateReq) (string, error) {
	now := time.Now().Unix()
	newNote := &reponote.Note{
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  int64(req.Basic.Privacy),
		Owner:    uid,
		CreateAt: now,
		UpdateAt: now,
	}

	var noteId int64
	err := s.dao.DB().TransactCtx(ctx, func(ctx context.Context, sess sqlx.Session) error {
		// 插入图片基础内容
		err := s.dao.NoteRepo.InsertTx(newNote, func(id, cnt int64) {
			noteId = id
		})(ctx, sess)

		if err != nil {
			logx.Errorf("note repo insert tx err: %v", err)
			return err
		}

		// 插入笔记资源数据
		var noteAssets = make([]*reponote.NoteAsset, 0, len(req.Images))
		for _, img := range req.Images {
			noteAssets = append(noteAssets, &reponote.NoteAsset{
				AssetKey:  img.FileId,
				AssetType: global.AssetTypeImage,
				NoteId:    noteId,
				CreateAt:  now,
			})
		}
		err = s.dao.NoteAssetRepo.BatchInsertTx(noteAssets)(ctx, sess)
		if err != nil {
			logx.Errorf("noteasset repo batch insert tx err: %v, noteid: %d", err, noteId)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("repo transact insert note err: %v, req: %+v, uid: %d", err, req, uid)
		return "", global.ErrInsertNoteFail
	}

	return s.NoteIdConfuser.Confuse(noteId), nil
}

func (s *Manage) Update(ctx context.Context, uid int64, req *mgtp.UpdateReq) error {
	now := time.Now().Unix()
	id := s.NoteIdConfuser.DeConfuse(req.NoteId)
	logx.Debugf("manage updating noteid: %d", id)
	queried, err := s.dao.NoteRepo.FindOne(ctx, id)
	if errors.Is(reponote.ErrNotFound, err) {
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

	newNote := &reponote.Note{
		Id:       id,
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  int64(req.Basic.Privacy),
		Owner:    queried.Owner,
		CreateAt: queried.CreateAt,
		UpdateAt: now,
	}

	// 开启事务执行
	err = s.dao.DB().TransactCtx(ctx, func(ctx context.Context, sess sqlx.Session) error {
		// 先更新基础信息
		err := s.dao.NoteRepo.UpdateTx(newNote)(ctx, sess)
		if err != nil {
			logx.Errorf("note repo update tx err: %v, noteid: %d", err, id)
			return err
		}

		oldAssets, err := s.dao.NoteAssetRepo.FindByNoteIdTx(ctx, sess, id)
		if err != nil && !errors.Is(reponote.ErrNotFound, err) {
			logx.Errorf("noteasset repo find err: %v, noteid: %d", err, id)
			return err
		}
		newAssetKeys := make([]string, 0, len(req.Images))
		for _, img := range req.Images {
			newAssetKeys = append(newAssetKeys, img.FileId)
		}

		// 随后删除旧资源
		err = s.dao.NoteAssetRepo.DeleteByNoteIdTx(id, newAssetKeys)(ctx, sess)
		if err != nil {
			logx.Errorf("noteasset repo delete tx err: %v, noteid: %d", err, id)
			return err
		}

		// 找出old和new的资源差异，只更新发生了变化的部分

		oldAssetMap := make(map[string]struct{})
		for _, old := range oldAssets {
			oldAssetMap[old.AssetKey] = struct{}{}
		}
		newAssets := make([]*reponote.NoteAsset, 0, len(req.Images))
		for _, img := range req.Images {
			if _, ok := oldAssetMap[img.FileId]; !ok {
				newAssets = append(newAssets, &reponote.NoteAsset{
					AssetKey:  img.FileId,
					AssetType: global.AssetTypeImage,
					NoteId:    id,
					CreateAt:  now,
				})
			}
		}

		if len(newAssets) == 0 {
			return nil
		}

		// 插入新的资源
		err = s.dao.NoteAssetRepo.BatchInsertTx(newAssets)(ctx, sess)
		if err != nil {
			logx.Errorf("noteasset repo batch insert tx err: %v, noteid: %d", err, id)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("repo transact update note err: %v, id: %d", err, id)
	}

	return nil
}

func (s *Manage) UploadAuth(ctx context.Context, req *mgtp.UploadAuthReq) (*mgtp.UploadAuthRes, error) {
	// 生成count个上传凭证
	fileId := s.Ctx.KeyGen.Gen()

	now := time.Now()
	currentTime := now.Unix()

	// 生成签名
	info, err := s.Signer.Sign(fileId, req.MimeType)
	if err != nil {
		logx.Errorf("upload auth sign err: %v, fileid: %s", err, fileId)
		return nil, global.ErrPermDenied.Msg("服务器签名失败")
	}

	res := mgtp.UploadAuthRes{
		FildIds:     fileId,
		CurrentTime: currentTime,
		ExpireTime:  info.ExpireAt.Unix(),
		UploadAddr:  s.Ctx.Config.Oss.Endpoint,
		Headers: mgtp.UploadAuthResHeaders{
			Auth:   info.Auth,
			Date:   info.Date,
			Sha256: info.Sha256,
			Token:  info.Token,
		},
	}

	return &res, nil
}

func (s *Manage) Delete(ctx context.Context, uid int64, req *mgtp.DeleteReq) error {
	id := s.NoteIdConfuser.DeConfuse(req.NoteId)
	if id <= 0 {
		return global.ErrNoteNotFound
	}

	logx.Debugf("manage updating noteid: %d", id)
	queried, err := s.dao.NoteRepo.FindOne(ctx, id)
	if errors.Is(reponote.ErrNotFound, err) {
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
	err = s.dao.DB().TransactCtx(ctx, func(ctx context.Context, sess sqlx.Session) error {
		err := s.dao.NoteRepo.DeleteTx(id)(ctx, sess)
		if err != nil {
			logx.Errorf("repo delete note basic tx err: %v, noteid: %d", err, id)
			return err
		}

		err = s.dao.NoteAssetRepo.DeleteByNoteIdTx(id, nil)(ctx, sess)
		if err != nil {
			logx.Errorf("repo delete note asset tx err: %v, noteid: %d", err, id)
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("repo delete note tx err: %v, noteid: %d", err, id)
		return global.ErrDeleteNoteFail
	}

	return nil
}

func (s *Manage) List(ctx context.Context, uid int64) (*mgtp.ListRes, error) {
	notes, err := s.dao.NoteRepo.ListByOwner(ctx, uid)
	if errors.Is(reponote.ErrNotFound, err) {
		return &mgtp.ListRes{}, nil
	}

	if err != nil {
		logx.Errorf("repo note list by owner err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	var noteIds = make([]int64, 0, len(notes))
	for _, note := range notes {
		noteIds = append(noteIds, note.Id)
	}

	// 获取资源信息
	noteAssets, err := s.dao.NoteAssetRepo.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, reponote.ErrNotFound) {
		logx.Errorf("repo note list by owner err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	// 组合notes和noteAssets
	var res mgtp.ListRes
	for _, note := range notes {
		item := &mgtp.ListResItem{
			NoteId:   s.NoteIdConfuser.Confuse(note.Id),
			Title:    note.Title,
			Desc:     note.Desc,
			Privacy:  note.Privacy,
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		for _, asset := range noteAssets {
			if note.Id == asset.NoteId {
				item.Images = append(item.Images, &mgtp.ListResItemImage{
					Url:  asset.AssetKey, // TODO 替换成oss能够访问的链接
					Type: int(asset.AssetType),
				})
			}
		}

		res.Items = append(res.Items, item)
	}

	return &res, nil
}

func (s *Manage) GetNote(ctx context.Context, uid int64, noteId string) (*mgtp.ListResItem, error) {
	nid := s.NoteIdConfuser.DeConfuse(noteId)
	if nid <= 0 {
		return nil, global.ErrNoteNotFound
	}

	note, err := s.dao.NoteRepo.FindOne(ctx, nid)
	if errors.Is(err, reponote.ErrNotFound) {
		return nil, global.ErrNoteNotFound
	}

	if err != nil {
		logx.Errorf("repo note find one err: %v, uid: %d", err, uid)
		return nil, global.ErrGetNoteFail
	}

	if note.Owner != uid {
		return nil, global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	assets, err := s.dao.NoteAssetRepo.FindByNoteIds(ctx, []int64{note.Id})
	if err != nil && !errors.Is(err, reponote.ErrNotFound) {
		logx.Errorf("repo note asset find by note ids err: %v, noteid: %d", err, note.Id)
		return nil, global.ErrGetNoteFail
	}

	var res = mgtp.ListResItem{
		NoteId:   noteId,
		Title:    note.Title,
		Desc:     note.Desc,
		Privacy:  note.Privacy,
		CreateAt: note.CreateAt,
		UpdateAt: note.UpdateAt,
	}

	for _, asset := range assets {
		res.Images = append(res.Images, &mgtp.ListResItemImage{
			Url:  asset.AssetKey, // TODO 替换oss
			Type: int(asset.AssetType),
		})
	}

	return &res, nil
}
