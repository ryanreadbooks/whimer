package noteevent

import (
	"context"

	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
	userdomain "github.com/ryanreadbooks/whimer/pilot/internal/domain/user"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	pkgnote "github.com/ryanreadbooks/whimer/note/pkg/event/note"
)

// 笔记事件处理
type Service struct {
	noteFeedAdapter     noterepo.NoteFeedAdapter
	noteSearchAdapter   noterepo.NoteSearchAdapter
	userServiceAdapter  userrepo.UserServiceAdapter
	systemNotifyService *systemnotify.DomainService
	userDomainService   *userdomain.DomainService
}

func NewService(
	noteFeedAdapter noterepo.NoteFeedAdapter,
	noteSearchAdapter noterepo.NoteSearchAdapter,
	userServiceAdapter userrepo.UserServiceAdapter,
	systemNotifyService *systemnotify.DomainService,
	userDomainService *userdomain.DomainService,
) *Service {
	return &Service{
		noteFeedAdapter:     noteFeedAdapter,
		noteSearchAdapter:   noteSearchAdapter,
		userServiceAdapter:  userServiceAdapter,
		systemNotifyService: systemNotifyService,
		userDomainService:   userDomainService,
	}
}

// 订阅笔记发布事件
//
// 执行如下操作
//  1. 笔记写入ES
//  2. 通知被At的人
//  3. 如果有被At的人 这些人写入最近联系人中
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

	// 1. 写入es
	err = s.noteSearchAdapter.AddNote(ctx, &entity.SearchNote{
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
	if err != nil {
		// log here
		xlog.Msgf("on note published search add note failed").Extras("note_id", ev.Note.Id).Errorx(ctx)
	}

	note, noteExt, err := s.noteFeedAdapter.GetNote(ctx, ev.Note.Id)
	if err != nil {
		return xerror.Wrapf(err, "on note published get note failed").WithCtx(ctx)
	}

	if len(noteExt.AtUsers) == 0 {
		return nil
	}

	atUids := make([]int64, 0, len(noteExt.AtUsers))
	for _, atUser := range noteExt.AtUsers {
		atUids = append(atUids, atUser.Uid)
	}

	atUids = xslice.Uniq(atUids)
	atUsers, err := s.userServiceAdapter.BatchGetUser(ctx, atUids)
	if err != nil {
		xlog.Msg("note creator user biz list users failed").Err(err).Errorx(ctx)
		return err
	}

	if len(atUsers) == 0 {
		xlog.Msg("user biz return empty at users").Errorx(ctx)
		return nil
	}

	validNoteAtUsers := make([]*mentionvo.AtUser, 0, len(atUsers))
	for _, atUser := range atUsers {
		if _, ok := atUsers[atUser.Uid]; ok {
			validNoteAtUsers = append(validNoteAtUsers, &mentionvo.AtUser{Uid: atUser.Uid, Nickname: atUser.Nickname})
		}
	}

	// 2. 通知
	s.notifyWhenAtUsers(ctx, note, validNoteAtUsers)

	// 3. 写入最近联系人
	s.appendRecentContacts(ctx, note.AuthorUid, validNoteAtUsers)

	return nil
}

func (s *Service) OnNoteDeleted(ctx context.Context, ev pkgnote.NoteDeletedEventData) error {
	return s.noteSearchAdapter.DeleteNote(ctx, notevo.NoteId(ev.Note.Id))
}

func (s *Service) notifyWhenAtUsers(ctx context.Context, note *entity.FeedNote, atUsers []*mentionvo.AtUser) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "pilot.noteevent.on_published.notify_at_users",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			err := s.systemNotifyService.NotifyAtUsersOnNote(ctx, &vo.NotifyAtUsersOnNoteParam{
				Uid:         note.AuthorUid,
				TargetUsers: atUsers,
				Content: &vo.NotifyAtUsersOnNoteParamContent{
					NoteDesc:  note.Desc,
					NoteId:    note.Id,
					SourceUid: note.AuthorUid,
				},
			})
			if err != nil {
				return xerror.Wrapf(err, "note creator notify biz failed to notify at users").WithCtx(ctx)
			}

			return nil
		},
	})
}

func (s *Service) appendRecentContacts(ctx context.Context, uid int64, atUsers []*mentionvo.AtUser) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "pilot.noteevent_on_published.append_recent_contacts",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			return s.userDomainService.AppendRecentContactsAtUser(ctx, uid, atUsers)
		},
	})
}
