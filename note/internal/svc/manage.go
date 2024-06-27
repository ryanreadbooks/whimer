package svc

import (
	"context"

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

func (s *Manage) Create(ctx context.Context, req *mgtyp.CreateReq) (string, error) {
	res, err := s.dao.NoteRepo.Insert(ctx, &note.Note{
		Title:   req.Basic.Title,
		Desc:    req.Basic.Desc,
		Privacy: int64(req.Basic.Privacy),
	})

	if err != nil {
		logx.Errorf("repo insert note err: %v, req: %+v", err, req)
		return "", global.ErrInsertNote
	}

	id, err := res.LastInsertId()
	if id <= 0 || err != nil {
		logx.Errorf("repo insert note err: %v, id: %d", err, id)
		return "", global.ErrInsertNote
	}

	return safety.Confuse(id), nil
}

func (s *Manage) Update(ctx context.Context, req *mgtyp.CreateReq) error {

	return nil
}
