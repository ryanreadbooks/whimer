package grpc

import (
	"context"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
)

type NoteCreatorServiceServer struct {
	notev1.UnimplementedNoteCreatorServiceServer

	Srv *srv.Service
}

func NewNoteAdminServiceServer(srv *srv.Service) *NoteCreatorServiceServer {
	return &NoteCreatorServiceServer{
		Srv: srv,
	}
}

func (s *NoteCreatorServiceServer) IsUserOwnNote(ctx context.Context, in *notev1.IsUserOwnNoteRequest) (
	*notev1.IsUserOwnNoteResponse, error,
) {
	owner, err := s.Srv.NoteCreatorSrv.GetNoteOwner(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsUserOwnNoteResponse{Uid: in.Uid, Result: owner == in.Uid}, nil
}

// 判断笔记是否存在
func (s *NoteCreatorServiceServer) IsNoteExist(ctx context.Context, in *notev1.IsNoteExistRequest) (
	*notev1.IsNoteExistResponse, error,
) {
	ok, err := s.Srv.NoteCreatorSrv.IsNoteExist(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsNoteExistResponse{Exist: ok}, nil
}

func convertToNoteAssets(in *notev1.CreateNoteRequest) ([]biz.CreateNoteRequestImage, *biz.CreateNoteRequestVideo) {
	var (
		images = make([]biz.CreateNoteRequestImage, 0, len(in.GetImages()))
		video  *biz.CreateNoteRequestVideo
	)
	if in.Basic.GetAssetType() == notev1.NoteAssetType_IMAGE {
		for _, img := range in.GetImages() {
			images = append(images, biz.CreateNoteRequestImage{
				FileId: img.GetFileId(),
				Width:  img.GetWidth(),
				Height: img.GetHeight(),
				Format: img.GetFormat(),
			})
		}
	} else if in.Basic.GetAssetType() == notev1.NoteAssetType_VIDEO {
		video = &biz.CreateNoteRequestVideo{
			FileId:       in.GetVideo().GetFileId(),
			TargetFileId: in.GetVideo().GetTargetFileId(),
			CoverFileId:  in.GetVideo().GetCoverFileId(),
		}
	}

	return images, video
}

func convertToNoteCreateReq(in *notev1.CreateNoteRequest) *biz.CreateNoteRequest {
	images, video := convertToNoteAssets(in)
	return &biz.CreateNoteRequest{
		Basic: biz.CreateNoteRequestBasic{
			Title:    in.Basic.Title,
			Desc:     in.Basic.Desc,
			Privacy:  model.Privacy(in.Basic.Privacy),
			NoteType: model.NoteType(in.Basic.AssetType),
		},
		Images:  images,
		TagIds:  in.GetTags().GetTagList(),
		AtUsers: model.AtUsersFromPb(in.GetAtUsers()),
		Video:   video,
	}
}

// 创建笔记
func (s *NoteCreatorServiceServer) CreateNote(ctx context.Context, in *notev1.CreateNoteRequest) (
	*notev1.CreateNoteResponse, error,
) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	req := convertToNoteCreateReq(in)
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.Basic.NoteType == model.AssetTypeVideo {
		if err := req.Video.Validate(); err != nil {
			return nil, err
		}
	}

	// service to create note
	noteId, err := s.Srv.NoteCreatorSrv.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return &notev1.CreateNoteResponse{NoteId: noteId}, nil
}

// 更新笔记
func (s *NoteCreatorServiceServer) UpdateNote(ctx context.Context, in *notev1.UpdateNoteRequest) (
	*notev1.UpdateNoteResponse, error,
) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	req := biz.UpdateNoteRequest{
		NoteId:            in.NoteId,
		CreateNoteRequest: *convertToNoteCreateReq(in.GetNote()),
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}
	// 更新时允许视频fileId为空 表示不更新视频资源

	err := s.Srv.NoteCreatorSrv.Update(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.UpdateNoteResponse{NoteId: req.NoteId}, nil
}

// 删除笔记
func (s *NoteCreatorServiceServer) DeleteNote(ctx context.Context, in *notev1.DeleteNoteRequest) (
	*notev1.DeleteNoteResponse, error,
) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	req := biz.DeleteNoteRequest{
		NoteId: in.NoteId,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Srv.NoteCreatorSrv.Delete(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.DeleteNoteResponse{}, nil
}

// 用于笔记作者获取笔记的详细信息
func (s *NoteCreatorServiceServer) GetNote(ctx context.Context, in *notev1.GetNoteRequest) (
	*notev1.GetNoteResponse, error,
) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if in.NoteId == 0 {
		return nil, global.ErrNoteNotFound
	}

	data, err := s.Srv.NoteCreatorSrv.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteResponse{Note: data.AsPb()}, nil
}

// 列出笔记
func (s *NoteCreatorServiceServer) ListNote(ctx context.Context, in *notev1.ListNoteRequest) (
	*notev1.ListNoteResponse, error,
) {
	data, nextPage, err := s.Srv.NoteCreatorSrv.List(ctx, in.Cursor, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.ListNoteResponse{
		Items:      items,
		NextCursor: nextPage.NextCursor,
		HasNext:    nextPage.HasNext,
	}, nil
}

func (s *NoteCreatorServiceServer) GetPostedCount(ctx context.Context, in *notev1.GetPostedCountRequest) (
	*notev1.GetPostedCountResponse, error,
) {
	cnt, err := s.Srv.NoteCreatorSrv.GetPostedCount(ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	return &notev1.GetPostedCountResponse{Count: cnt}, nil
}

func (s *NoteCreatorServiceServer) PageListNote(ctx context.Context,
	in *notev1.PageListNoteRequest,
) (*notev1.PageListNoteResponse, error) {
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Count >= 20 {
		in.Count = 20
	}
	data, total, err := s.Srv.NoteCreatorSrv.PageList(ctx, in.Page, in.Count, in.LifeCycleState)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.PageListNoteResponse{
		Items: items,
		Total: total,
	}, nil
}

func (s *NoteCreatorServiceServer) AddTag(ctx context.Context, in *notev1.AddTagRequest) (
	*notev1.AddTagResponse, error,
) {
	if in.Name == "" {
		return nil, global.ErrArgs.Msg("标签名为空")
	}

	id, err := s.Srv.NoteCreatorSrv.AddTag(ctx, in.Name)
	if err != nil {
		return nil, err
	}

	return &notev1.AddTagResponse{Id: id}, nil
}
