package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
	reponote "github.com/ryanreadbooks/whimer/note/internal/repo/note"
	notetyp "github.com/ryanreadbooks/whimer/note/internal/types"
	mgtyp "github.com/ryanreadbooks/whimer/note/internal/types/manage"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	noteIdConfuserSalt = "0x7c00:noteIdConfuser:.$35%io"
)

type Manage struct {
	Ctx            *ServiceContext
	dao            *repo.Dao
	NoteIdConfuser *safety.Confuser
}

func NewManage(repo *repo.Dao) *Manage {
	return &Manage{
		dao:            repo,
		NoteIdConfuser: safety.NewConfuser(noteIdConfuserSalt, 24),
	}
}

func (s *Manage) Get(ctx context.Context, uid int64, noteId string) error {

	return nil
}

func (s *Manage) Create(ctx context.Context, uid int64, req *mgtyp.CreateReq) (string, error) {
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

func (s *Manage) Update(ctx context.Context, uid int64, req *mgtyp.UpdateReq) error {
	now := time.Now().Unix()
	id := s.NoteIdConfuser.DeConfuse(req.NoteId)
	logx.Debugf("manage updating noteid: %d", id)
	queried, err := s.dao.NoteRepo.FindOne(ctx, id)
	if errors.Is(reponote.ErrNotFound, err) {
		return global.ErrUpdateNoteNotFound
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

func (s *Manage) UploadAuth(ctx context.Context, req *notetyp.UploadAuthReq) (*notetyp.UploadAuthRes, error) {
	// 生成count个上传凭证
	var fileIds []string = make([]string, 0, req.Count)
	for i := 0; i < req.Count; i++ {
		key := s.Ctx.KeyGen.Gen()
		fileIds = append(fileIds, key)
	}

	currentTime := time.Now().Unix()

	res := notetyp.UploadAuthRes{
		FildIds:     fileIds,
		CurrentTime: currentTime,
	}

	return &res, nil
}
