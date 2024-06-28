package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/safety"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
	"github.com/ryanreadbooks/whimer/note/internal/repo/note"
	mgtyp "github.com/ryanreadbooks/whimer/note/internal/types/manage"
	"github.com/zeromicro/go-zero/core/logx"
)

type Manage struct {
	dao *repo.Dao
}

func NewManage(repo *repo.Dao) *Manage {
	return &Manage{
		dao: repo,
	}
}

func (s *Manage) Create(ctx context.Context, uid int64, req *mgtyp.CreateReq) (string, error) {
	now := time.Now().Unix()
	newNote := &note.Note{
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  int64(req.Basic.Privacy),
		Owner:    uid,
		CreateAt: now,
		UpdateAt: now,
	}

	var noteId int64
	opInsertNote := s.dao.NoteRepo.InsertTx(newNote, func(id, cnt int64) {
		noteId = id
	})

	// 开启事务插入图片内容
	// TODO 插入笔记的图片等资源
	err := s.dao.Transact(ctx, opInsertNote)

	if err != nil {
		logx.Errorf("repo transact insert note err: %v, req: %+v, uid: %d", err, req, uid)
		return "", global.ErrInsertNoteFail
	}

	return safety.Confuse(noteId), nil
}

func (s *Manage) Update(ctx context.Context, uid int64, req *mgtyp.UpdateReq) error {
	now := time.Now().Unix()
	id := safety.DeConfuse(req.NoteId)
	queried, err := s.dao.NoteRepo.FindOne(ctx, id)
	if errors.Is(note.ErrNotFound, err) {
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

	newNote := &note.Note{
		Id:       id,
		Title:    req.Basic.Title,
		Desc:     req.Basic.Desc,
		Privacy:  int64(req.Basic.Privacy),
		Owner:    queried.Owner,
		CreateAt: queried.CreateAt,
		UpdateAt: now,
	}

	// 开启事务更新笔记
	opUpdateNote := s.dao.NoteRepo.UpdateTx(newNote)
	// TODO 更新笔记的图片等资源
	err = s.dao.Transact(ctx, opUpdateNote)
	if err != nil {
		logx.Errorf("repo transact update note err: %v, req: %+v, uid: %d", err, req, uid)
		return global.ErrUpdateNoteFail
	}

	return nil
}
