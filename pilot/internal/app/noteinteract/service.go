package noteinteract

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract/dto"
	commentrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
)

// 只负责业务逻辑编排
type Service struct {
	noteLikesAdapter noterepo.NoteLikesAdapter
	commentAdapter   commentrepo.CommentAdapter
}

func NewService(
	noteLikesAdapter noterepo.NoteLikesAdapter,
	commentAdapter commentrepo.CommentAdapter,
) *Service {
	return &Service{
		noteLikesAdapter: noteLikesAdapter,
		commentAdapter:   commentAdapter,
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

	return nil
}

// 获取笔记点赞数量
func (s *Service) GetLikeCount(ctx context.Context, noteId int64) (int64, error) {
	cnt, err := s.noteLikesAdapter.GetLikeCount(ctx, noteId)
	return cnt, err
}
