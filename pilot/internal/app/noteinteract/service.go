package noteinteract

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract/dto"
	commentrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

// 只负责业务逻辑编排
type Service struct {
	noteLikesAdapter    noterepo.NoteLikesAdapter
	noteFeedAdapter     noterepo.NoteFeedAdapter
	commentAdapter      commentrepo.CommentAdapter
	systemNotifyService *systemnotify.DomainService
}

func NewService(
	noteLikesAdapter noterepo.NoteLikesAdapter,
	noteFeedAdapter noterepo.NoteFeedAdapter,
	commentAdapter commentrepo.CommentAdapter,
	systemNotifyService *systemnotify.DomainService,
) *Service {
	return &Service{
		noteLikesAdapter:    noteLikesAdapter,
		noteFeedAdapter:     noteFeedAdapter,
		commentAdapter:      commentAdapter,
		systemNotifyService: systemNotifyService,
	}
}

// 点赞笔记
func (s *Service) LikeNote(ctx context.Context, cmd *dto.LikeNoteCommand) error {
	uid := metadata.Uid(ctx)
	err := s.noteLikesAdapter.LikeNote(ctx, &noterepo.LikeNoteParams{
		Uid:    uid,
		NoteId: cmd.NoteId.Int64(),
		Action: cmd.Action,
	})
	if err != nil {
		return err
	}

	if cmd.Action == notevo.LikeActionDo {
		s.notifyUserNoteLiked(ctx, uid, cmd.NoteId.Int64())
	}

	return nil
}

// 获取笔记点赞数量
func (s *Service) GetLikeCount(ctx context.Context, noteId int64) (int64, error) {
	cnt, err := s.noteLikesAdapter.GetLikeCount(ctx, noteId)
	return cnt, err
}

func (s *Service) notifyUserNoteLiked(ctx context.Context, uid, noteId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "pilot.noteinteract.notify_user",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			// 拿到笔记的作者
			author, err := s.noteFeedAdapter.GetNoteAuthorUid(ctx, noteId)
			if err != nil {
				return xerror.Wrapf(err, "failed to get note author").WithExtra("note_id", noteId).WithCtx(ctx)
			}

			return s.systemNotifyService.NotifyUserLikesOnNote(ctx, uid, author, &vo.NotifyLikesOnNoteParam{
				NoteId: notevo.NoteId(noteId),
			})
		},
	})
}
