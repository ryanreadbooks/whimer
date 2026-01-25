package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/note/convert"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

// implement domain/note/repository/creator
var _ repository.NoteCreatorAdapter = &CreatorAdapterImpl{}

type CreatorAdapterImpl struct {
	noteCreatorCli notev1.NoteCreatorServiceClient
	noteFeedCli    notev1.NoteFeedServiceClient
	searchCli      searchv1.SearchServiceClient
	searchDocCli   searchv1.DocumentServiceClient
}

func NewCreatorAdapterImpl(
	noteCreatorCli notev1.NoteCreatorServiceClient,
	noteFeedCli notev1.NoteFeedServiceClient,
	searchCli searchv1.SearchServiceClient,
	searchDocCli searchv1.DocumentServiceClient,
) *CreatorAdapterImpl {
	return &CreatorAdapterImpl{
		noteCreatorCli: noteCreatorCli,
		noteFeedCli:    noteFeedCli,
		searchCli:      searchCli,
		searchDocCli:   searchDocCli,
	}
}

// 获取笔记
func (a *CreatorAdapterImpl) GetNote(ctx context.Context, noteId int64) (*entity.CreatorNote, error) {
	req := &notev1.GetNoteRequest{NoteId: noteId}
	resp, err := a.noteCreatorCli.GetNote(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return convert.PbNoteToCreatorEntity(resp.GetNote()), nil
}

// 新建笔记
func (a *CreatorAdapterImpl) CreateNote(ctx context.Context, params *entity.CreateNoteParams) (int64, error) {
	req := convert.EntityCreateNoteParamsAsPb(params)
	resp, err := a.noteCreatorCli.CreateNote(ctx, req)
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetNoteId(), nil
}

// 更新笔记
func (a *CreatorAdapterImpl) UpdateNote(ctx context.Context, params *entity.UpdateNoteParams) (int64, error) {
	req := &notev1.UpdateNoteRequest{
		NoteId: params.NoteId,
		Note:   convert.EntityCreateNoteParamsAsPb(&params.CreateNoteParams),
	}
	resp, err := a.noteCreatorCli.UpdateNote(ctx, req)
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetNoteId(), nil
}

// 删除笔记
func (a *CreatorAdapterImpl) DeleteNote(ctx context.Context, noteId int64) error {
	req := &notev1.DeleteNoteRequest{NoteId: noteId}
	_, err := a.noteCreatorCli.DeleteNote(ctx, req)
	if err != nil {
		return xerror.Wrap(err)
	}
	return nil
}

// 分页获取笔记列表
func (a *CreatorAdapterImpl) PageListNotes(ctx context.Context, params *entity.PageListNotesParams) (*entity.PageListNotesResult, error) {
	req := &notev1.PageListNoteRequest{
		Page:           params.Page,
		Count:          params.Count,
		LifeCycleState: convert.NoteStatusToLifeCycleState(params.Status),
	}
	resp, err := a.noteCreatorCli.PageListNote(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &entity.PageListNotesResult{
		Total: resp.GetTotal(),
		Items: convert.BatchPbNotesToCreatorEntities(resp.GetItems()),
	}, nil
}

// 新增标签
func (a *CreatorAdapterImpl) AddTag(ctx context.Context, name string) (int64, error) {
	resp, err := a.noteCreatorCli.AddTag(ctx, &notev1.AddTagRequest{Name: name})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	// 异步同步search
	tagId := resp.GetId()
	a.asyncPutTagToSearch(ctx, tagId)

	return resp.GetId(), nil
}

// 搜索标签
func (a *CreatorAdapterImpl) SearchTags(ctx context.Context, name string) ([]*entity.SearchedNoteTag, error) {
	resp, err := a.searchCli.SearchNoteTags(ctx,
		&searchv1.SearchNoteTagsRequest{Text: name})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	result := make([]*entity.SearchedNoteTag, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		if item == nil {
			continue
		}
		result = append(result, &entity.SearchedNoteTag{
			Id:   item.GetId(),
			Name: item.GetName(),
		})
	}

	return result, nil
}

// 获取标签
func (a *CreatorAdapterImpl) GetTag(ctx context.Context, tagId int64) (*entity.NoteTag, error) {
	resp, err := a.noteFeedCli.GetTagInfo(ctx, &notev1.GetTagInfoRequest{Id: tagId})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return &entity.NoteTag{
		Id:   resp.GetTag().GetId(),
		Name: resp.GetTag().GetName(),
	}, nil
}

func (a *CreatorAdapterImpl) asyncPutTagToSearch(ctx context.Context, tagId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "pilot.adapter.sync_tag",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			newTag, err := a.noteFeedCli.GetTagInfo(ctx,
				&notev1.GetTagInfoRequest{Id: tagId})
			if err != nil {
				return xerror.Wrap(err)
			}

			tid := vo.TagId(newTag.GetTag().GetId()).String()
			_, err = a.searchDocCli.BatchAddNoteTag(ctx,
				&searchv1.BatchAddNoteTagRequest{
					NoteTags: []*searchv1.NoteTag{{
						Id:    tid,
						Name:  newTag.GetTag().GetName(),
						Ctime: newTag.GetTag().GetCtime(),
					}},
				},
			)
			if err != nil {
				return xerror.Wrap(err)
			}

			return nil
		},
	})
}
