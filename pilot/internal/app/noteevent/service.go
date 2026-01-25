package noteevent

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	pkgnote "github.com/ryanreadbooks/whimer/note/pkg/event/note"
)

// 笔记事件处理
type Service struct {
	noteSearchAdapter  noterepo.NoteSearchAdapter
	userServiceAdapter userrepo.UserServiceAdapter
}

func NewService(
	noteSearchAdapter noterepo.NoteSearchAdapter,
	userServiceAdapter userrepo.UserServiceAdapter,
) *Service {
	return &Service{
		noteSearchAdapter:  noteSearchAdapter,
		userServiceAdapter: userServiceAdapter,
	}
}

func (s *Service) OnNotePublished(ctx context.Context, ev pkgnote.NotePublishedEventData) error {
	// get user nickname
	username := ""
	user, err := s.userServiceAdapter.GetUser(ctx, ev.Note.Owner)
	if err != nil {
		xlog.Msg("note event service on note published failed to get user nickname").Errorx(ctx)
	} else {
		username = user.Nickname
	}

	tagList := []*entity.SearchedNoteTag{}
	for _, t := range ev.Note.Tags {
		tagList = append(tagList, &entity.SearchedNoteTag{
			Id:    t.Tid,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	var assetType notevo.AssetType
	switch ev.Note.Type {
	case "image":
		assetType = notevo.AssetTypeImage
	case "video":
		assetType = notevo.AssetTypeVideo
	}

	s.noteSearchAdapter.AddNote(ctx, &entity.SearchNote{
		NoteId:         notevo.NoteId(ev.Note.Id),
		Title:          ev.Note.Title,
		Desc:           ev.Note.Desc,
		CreateAt:       ev.Note.Ctime,
		UpdateAt:       ev.Note.Utime,
		AuthorUid:      ev.Note.Owner,
		AuthorNickname: username,
		TagList:        tagList,
		AssetType:      assetType,
		Visibility:     notevo.VisibilityPublic,
	})

	return nil
}

func (s *Service) OnNoteDeleted(ctx context.Context, ev pkgnote.NoteDeletedEventData) error {
	return s.noteSearchAdapter.DeleteNote(ctx, notevo.NoteId(ev.Note.Id))
}
