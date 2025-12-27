package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv/procedure"
)

// type alias for convenience
type (
	CreateNoteRequest = biz.CreateNoteRequest
	UpdateNoteRequest = biz.UpdateNoteRequest
	DeleteNoteRequest = biz.DeleteNoteRequest
)

type NoteCreatorSrv struct {
	parent *Service

	biz              *biz.Biz
	noteBiz          *biz.NoteBiz
	noteProcedureBiz *biz.NoteProcedureBiz
	noteCreatorBiz   *biz.NoteCreatorBiz
	noteInteractBiz  *biz.NoteInteractBiz

	procedureMgr *procedure.Manager
}

func NewNoteCreatorSrv(p *Service, biz *biz.Biz, procedureMgr *procedure.Manager) *NoteCreatorSrv {
	return &NoteCreatorSrv{
		parent:           p,
		biz:              biz,
		noteBiz:          biz.Note,
		noteProcedureBiz: biz.Procedure,
		noteCreatorBiz:   biz.Creator,
		noteInteractBiz:  biz.Interact,
		procedureMgr:     procedureMgr,
	}
}

// 新建笔记
func (s *NoteCreatorSrv) Create(ctx context.Context, req *CreateNoteRequest) (int64, error) {
	// check tag ids
	tagIds := req.TagIds
	if len(tagIds) > 0 {
		reqTag, err := s.noteBiz.BatchGetTag(ctx, tagIds)
		if err != nil {
			return 0, xerror.Wrapf(err, "srv creator batch get tag failed").WithCtx(ctx)
		}

		if len(reqTag) != len(tagIds) {
			return 0, global.ErrTagNotFound
		}
	}

	var (
		newNote *model.Note
		proceed func(ctx context.Context) bool
	)

	// create note
	err := s.biz.Tx(ctx, func(ctx context.Context) error {
		var errTx error
		newNote, errTx = s.noteCreatorBiz.CreateNote(ctx, req)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator create note failed").WithCtx(ctx)
		}

		// 初始化流程记录
		proceed, errTx = s.procedureMgr.RunPipeline(ctx, newNote, procedure.StartAtAssetProcess())
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator init procedure failed").WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator create note tx failed").WithCtx(ctx)
	}

	// 继续执行流程任务
	_ = proceed(ctx)

	return newNote.NoteId, nil
}

// 更新笔记
func (s *NoteCreatorSrv) Update(ctx context.Context, req *UpdateNoteRequest) error {
	var (
		newNote *model.Note
		proceed func(ctx context.Context) bool
	)

	err := s.biz.Tx(ctx, func(ctx context.Context) error {
		var errTx error
		newNote, errTx = s.noteCreatorBiz.UpdateNote(ctx, req)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator update note failed").WithCtx(ctx)
		}

		startAt := procedure.StartAtAssetProcess()
		if newNote.State != model.NoteStateInit {
			startAt = procedure.StartAtAudit()
		}

		proceed, errTx = s.procedureMgr.RunPipeline(ctx, newNote, startAt)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator init procedure failed").WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "srv creator update note tx failed").WithCtx(ctx)
	}

	// 继续后续流程
	_ = proceed(ctx)

	return nil
}

// 删除笔记
func (s *NoteCreatorSrv) Delete(ctx context.Context, req *DeleteNoteRequest) error {
	// TODO 停掉正在进行的procedure处理
	err := s.biz.Tx(ctx, func(ctx context.Context) error {
		return s.noteCreatorBiz.DeleteNote(ctx, req)
	})
	if err != nil {
		return xerror.Wrapf(err, "srv creator delete note tx failed").WithCtx(ctx)
	}

	return nil
}

// 列出某用户所有笔记
func (s *NoteCreatorSrv) List(ctx context.Context, cursor int64, count int32) (*model.Notes, model.PageResult, error) {
	resp, nextPage, err := s.noteCreatorBiz.PageListNoteWithCursor(ctx, cursor, count)
	if err != nil {
		return nil, nextPage, xerror.Wrapf(err, "srv creator cursor list note failed").WithCtx(ctx)
	}

	resp, _ = s.noteInteractBiz.AssignNoteLikes(ctx, resp)
	resp, _ = s.noteInteractBiz.AssignNoteReplies(ctx, resp)

	return resp, nextPage, nil
}

func (s *NoteCreatorSrv) PageList(ctx context.Context, page, count int32, lcState notev1.NoteLifeCycleState) (*model.Notes, int64, error) {
	resp, total, err := s.noteCreatorBiz.PageListNote(ctx, page, count, lcState)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "srv creator page list note failed").WithCtx(ctx)
	}

	resp, _ = s.noteInteractBiz.AssignNoteLikes(ctx, resp)
	resp, _ = s.noteInteractBiz.AssignNoteReplies(ctx, resp)

	return resp, total, nil
}

// 用于笔记作者获取笔记的详细信息
func (s *NoteCreatorSrv) GetNote(ctx context.Context, noteId int64) (*model.Note, error) {
	note, err := s.noteCreatorBiz.CreatorGetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "srv creator get note failed").WithCtx(ctx)
	}

	resp, _ := s.noteInteractBiz.AssignNoteLikes(ctx, &model.Notes{Items: []*model.Note{note}})
	resp, _ = s.noteInteractBiz.AssignNoteReplies(ctx, resp)

	return resp.Items[0], nil
}

// 获取笔记作者
func (s *NoteCreatorSrv) GetNoteOwner(ctx context.Context, noteId int64) (int64, error) {
	return s.noteBiz.GetNoteOwner(ctx, noteId)
}

// 判断笔记是否存在
func (s *NoteCreatorSrv) IsNoteExist(ctx context.Context, noteId int64) (bool, error) {
	return s.noteBiz.IsNoteExist(ctx, noteId)
}

// 获取用户发布的笔记数量
func (s *NoteCreatorSrv) GetPostedCount(ctx context.Context, uid int64) (int64, error) {
	cnt, err := s.noteCreatorBiz.GetUserPostedCount(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator get count failed").WithCtx(ctx)
	}

	return cnt, nil
}

// 用户添加标签
func (s *NoteCreatorSrv) AddTag(ctx context.Context, name string) (int64, error) {
	id, err := s.noteCreatorBiz.AddTag(ctx, name)
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator add tag failed").WithCtx(ctx)
	}

	return id, nil
}
