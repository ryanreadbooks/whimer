package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
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

	procedureMgr *ProcedureManager
}

func NewNoteCreatorSrv(p *Service, biz *biz.Biz, procedureMgr *ProcedureManager) *NoteCreatorSrv {
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

	var newNote *model.Note

	// create note
	err := s.biz.Tx(ctx, func(ctx context.Context) error {
		var errTx error
		newNote, errTx = s.noteCreatorBiz.CreateNote(ctx, req)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator create note failed").WithCtx(ctx)
		}

		// 初始化流程记录
		errTx = s.procedureMgr.Create(ctx, newNote, model.ProcedureTypeAssetProcess)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator init procedure failed").WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator tx failed").WithCtx(ctx)
	}

	// 执行流程任务
	newTaskId, err := s.procedureMgr.Execute(ctx, newNote)
	if err != nil {
		// 此处笔记已入库 只是调度任务失败 后台重试不返回错误 此处仅打日志 + 打点告警
		xlog.Msg("srv creator execute procedure failed").
			Err(err).
			Extras("note_id", newNote.NoteId).
			Errorx(ctx)
	} else {
		// 确认流程（回填taskId）
		err = s.procedureMgr.Confirm(ctx, newNote.NoteId, newTaskId, model.ProcedureTypeAssetProcess)
		if err != nil {
			xlog.Msg("srv creator confirm procedure failed").
				Err(err).
				Extras("note_id", newNote.NoteId).
				Extras("taskId", newTaskId).
				Errorx(ctx)
		}
	}

	return newNote.NoteId, nil
}

// 更新笔记
func (s *NoteCreatorSrv) Update(ctx context.Context, req *UpdateNoteRequest) error {
	err := s.biz.Tx(ctx, func(ctx context.Context) error {
		err := s.noteCreatorBiz.UpdateNote(ctx, req)
		if err != nil {
			return xerror.Wrapf(err, "srv creator update note failed").WithCtx(ctx)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// TODO 重新走发布流程

	return nil
}

// 删除笔记
func (s *NoteCreatorSrv) Delete(ctx context.Context, req *DeleteNoteRequest) error {
	return s.noteCreatorBiz.DeleteNote(ctx, req)
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

func (s *NoteCreatorSrv) PageList(ctx context.Context, page, count int32) (*model.Notes, int64, error) {
	resp, total, err := s.noteCreatorBiz.PageListNote(ctx, page, count)
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
