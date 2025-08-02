package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	"golang.org/x/sync/errgroup"
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
func (s *NoteCreatorSrv) Create(ctx context.Context, req *model.CreateNoteRequest) (int64, error) {
	return s.noteCreatorBiz.CreateNote(ctx, req)
}

// 更新笔记
func (s *NoteCreatorSrv) Update(ctx context.Context, req *model.UpdateNoteRequest) error {
	return s.noteCreatorBiz.UpdateNote(ctx, req)
}

// 获取上传凭证
//
// Deprecated: UploadAuth is deprecated
func (s *NoteCreatorSrv) UploadAuth(ctx context.Context, req *model.UploadAuthRequest) (*model.UploadAuthResponse, error) {
	return s.noteCreatorBiz.GetUploadAuth(ctx, req)
}

// 批量获取上传凭证
//
// Deprecated: BatchGetUploadAuth is deprecated
func (s *NoteCreatorSrv) BatchGetUploadAuth(ctx context.Context,
	req *model.UploadAuthRequest) ([]*model.UploadAuthResponse, error) {

	eg, ctx := errgroup.WithContext(ctx)
	var resps = make([]*model.UploadAuthResponse, req.Count)
	for i := range req.Count {
		eg.Go(func() error {
			resp, err := s.noteCreatorBiz.GetUploadAuth(ctx, req)
			if err != nil {
				return xerror.Wrapf(err, "get upload auth failed").WithExtra("req", req)
			}

			resps[i] = resp
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return resps, nil
}

// 获取临时上传凭证
func (s *NoteCreatorSrv) BatchGetUploadAuthSTS(ctx context.Context,
	req *model.UploadAuthRequest) (*model.UploadAuthSTSResponse, error) {
	res, err := s.noteCreatorBiz.GetUploadAuthSTS(ctx, &model.UploadAuthRequest{
		Resource: req.Resource,
		Source:   req.Source,
		Count:    req.Count,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "srv creator get upload auth sts failed").WithCtx(ctx)
	}

	return res, nil
}

// 删除笔记
func (s *NoteCreatorSrv) Delete(ctx context.Context, req *model.DeleteNoteRequest) error {
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
	note, err := s.noteCreatorBiz.GetNote(ctx, noteId)
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
	cnt, err := infra.Dao().NoteDao.GetPostedCountByOwner(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "srv creator get posted count failed").
			WithExtra("uid", uid).
			WithCtx(ctx)
	}

	return cnt, nil
}
