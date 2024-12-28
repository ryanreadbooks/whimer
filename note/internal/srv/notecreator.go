package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type NoteCreatorSrv struct {
	parent *Service

	noteBiz         biz.NoteBiz
	noteCreatorBiz  biz.NoteCreatorBiz
	noteInteractBiz biz.NoteInteractBiz
}

func NewNoteCreatorSrv(p *Service, biz biz.Biz) *NoteCreatorSrv {
	return &NoteCreatorSrv{
		parent:          p,
		noteBiz:         biz.Note,
		noteCreatorBiz:  biz.Creator,
		noteInteractBiz: biz.Interact,
	}
}

// 新建笔记
func (s *NoteCreatorSrv) Create(ctx context.Context, req *model.CreateNoteRequest) (uint64, error) {
	return s.noteCreatorBiz.CreatorCreateNote(ctx, req)
}

// 更新笔记
func (s *NoteCreatorSrv) Update(ctx context.Context, req *model.UpdateNoteRequest) error {
	return s.noteCreatorBiz.CreatorUpdateNote(ctx, req)
}

// 获取上传凭证
func (s *NoteCreatorSrv) UploadAuth(ctx context.Context, req *model.UploadAuthRequest) (*model.UploadAuthResponse, error) {
	return s.noteCreatorBiz.CreatorGetUploadAuth(ctx, req)
}

// 删除笔记
func (s *NoteCreatorSrv) Delete(ctx context.Context, req *model.DeleteNoteRequest) error {
	return s.noteCreatorBiz.CreatorDeleteNote(ctx, req)
}

// 列出某用户所有笔记
func (s *NoteCreatorSrv) List(ctx context.Context, cursor uint64, count int32) (*model.Notes, model.PageResult, error) {
	resp, nextPage, err := s.noteCreatorBiz.CreatorPageListNote(ctx, cursor, count)
	if err != nil {
		return nil, nextPage, xerror.Wrapf(err, "srv creator list note failed").WithCtx(ctx)
	}

	resp, _ = s.noteInteractBiz.AssignNoteLikes(ctx, resp)
	resp, _ = s.noteInteractBiz.AssignNoteReplies(ctx, resp)

	return resp, nextPage, nil
}

// 用于笔记作者获取笔记的详细信息
func (s *NoteCreatorSrv) GetNote(ctx context.Context, noteId uint64) (*model.Note, error) {
	note, err := s.noteCreatorBiz.CreatorGetNote(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "srv creator get note failed").WithCtx(ctx)
	}

	resp, _ := s.noteInteractBiz.AssignNoteLikes(ctx, &model.Notes{Items: []*model.Note{note}})
	resp, _ = s.noteInteractBiz.AssignNoteReplies(ctx, resp)

	return resp.Items[0], nil
}

// 获取笔记作者
func (s *NoteCreatorSrv) GetNoteOwner(ctx context.Context, noteId uint64) (uint64, error) {
	return s.noteBiz.GetNoteOwner(ctx, noteId)
}

// 判断笔记是否存在
func (s *NoteCreatorSrv) IsNoteExist(ctx context.Context, noteId uint64) (bool, error) {
	return s.noteBiz.IsNoteExist(ctx, noteId)
}

// 获取用户发布的笔记数量
func (s *NoteCreatorSrv) GetPostedCount(ctx context.Context, uid uint64) (uint64, error) {
	cnt, err := infra.Dao().NoteDao.GetPostedCountByOwner(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator get posted count failed").
			WithExtra("uid", uid).
			WithCtx(ctx)
	}

	return cnt, nil
}
