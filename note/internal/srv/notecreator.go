package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv/assetprocess"
)

// type alias for convenience
type (
	CreateNoteRequest = biz.CreateNoteRequest
	UpdateNoteRequest = biz.UpdateNoteRequest
	DeleteNoteRequest = biz.DeleteNoteRequest
)

type NoteCreatorSrv struct {
	parent *Service

	biz              biz.Biz
	noteBiz          biz.NoteBiz
	noteProcedureBiz biz.NoteProcedureBiz
	noteCreatorBiz   biz.NoteCreatorBiz
	noteInteractBiz  biz.NoteInteractBiz
}

func NewNoteCreatorSrv(p *Service, biz biz.Biz) *NoteCreatorSrv {
	return &NoteCreatorSrv{
		parent:           p,
		biz:              biz,
		noteBiz:          biz.Note,
		noteProcedureBiz: biz.Procedure,
		noteCreatorBiz:   biz.Creator,
		noteInteractBiz:  biz.Interact,
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

		errTx = s.beforeEnterAssetProcessFlow(ctx, newNote)
		if errTx != nil {
			return xerror.Wrapf(errTx, "srv creator before enter publish flow failed").WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator tx failed").WithCtx(ctx)
	}

	// 注册调度任务
	newTaskId, err := s.enterPublishFlow(ctx, newNote)
	if err != nil {
		// 此处笔记已入库 只是调度任务失败 后台重试不返回错误 此处仅打日志 + 打点告警
		xlog.Msg("srv creator enter publish flow failed").
			Err(err).
			Extras("note_id", newNote.NoteId).
			Errorx(ctx)
	} else {
		// 回调taskId失败同样后台重试兜底 此处仅打日志
		err = s.afterEnterAssetProcessFlow(ctx, newNote, newTaskId)
		if err != nil {
			xlog.Msg("srv creator after enter publish flow failed").
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

// 发起远程调度任务注册前 先本地写表
func (s *NoteCreatorSrv) beforeEnterAssetProcessFlow(ctx context.Context, note *model.Note) error {
	// taskId 先留空后续再填充
	err := s.noteProcedureBiz.CreateRecord(ctx, &biz.CreateProcedureRecordReq{
		NoteId:      note.NoteId,
		Protype:     model.ProcedureTypeAssetProcess,
		TaskId:      "",
		MaxRetryCnt: 3,
	})
	if err != nil {
		return xerror.Wrapf(err, "srv creator create process record failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// 标记开始处理
	err = s.noteCreatorBiz.SetNoteStateProcessing(ctx, note.NoteId)
	if err != nil {
		return xerror.Wrapf(err, "srv creator set note state processing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return nil
}

func (s *NoteCreatorSrv) afterEnterAssetProcessFlow(ctx context.Context, note *model.Note, taskId string) error {
	// taskId 回填 optional
	err := s.noteProcedureBiz.UpdateTaskId(ctx, note.NoteId, model.ProcedureTypeAssetProcess, taskId)
	if err != nil {
		return xerror.Wrapf(err, "srv creator update process record task id failed").
			WithExtras("note_id", note.NoteId, "taskId", taskId).
			WithCtx(ctx)
	}

	return nil
}

// 进入发布流程
func (s *NoteCreatorSrv) enterPublishFlow(ctx context.Context, note *model.Note) (string, error) {
	// 调度处理笔记资源
	assetProcessor := assetprocess.NewProcessor(note.Type, s.biz)
	taskId, err := assetProcessor.Process(ctx, note)
	if err != nil {
		return "", xerror.Wrapf(err, "srv creator enter publish flow failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return taskId, nil
}
