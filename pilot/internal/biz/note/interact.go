package note

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/note/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"

	"golang.org/x/sync/errgroup"
)

// AssignNoteExtra 设置笔记的额外信息（点赞、评论状态）
func (b *Biz) AssignNoteExtra(ctx context.Context, notes []*imodel.AdminNoteItem) {
	if len(notes) == 0 {
		return
	}

	var (
		noteIds      = make([]int64, 0, len(notes))
		oidLiked     = make(map[int64]bool)
		oidCommented = make(map[int64]bool)
		uid          = metadata.Uid(ctx)
		eg           errgroup.Group
	)

	for _, n := range notes {
		noteIds = append(noteIds, int64(n.NoteId))
	}

	eg.Go(func() error {
		return recovery.Do(func() error {
			mappings := make(map[int64]*notev1.NoteIdList)
			mappings[uid] = &notev1.NoteIdList{NoteIds: noteIds}

			resp, err := dep.NoteInteractServer().BatchCheckUserLikeStatus(ctx,
				&notev1.BatchCheckUserLikeStatusRequest{Mappings: mappings})
			if err != nil {
				return xerror.Wrapf(err, "failed to get user like status").WithCtx(ctx)
			}

			pairs := resp.GetResults()
			for _, likedInfo := range pairs[uid].GetList() {
				oidLiked[likedInfo.NoteId] = likedInfo.Liked
			}

			for _, note := range notes {
				noteId := int64(note.NoteId)
				note.Interact.Liked = oidLiked[noteId]
			}

			return nil
		})
	})

	eg.Go(func() error {
		return recovery.Do(func() error {
			commentMappings := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
			commentMappings[uid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{Oids: noteIds}

			resp, err := dep.Commenter().BatchCheckUserOnObject(ctx,
				&commentv1.BatchCheckUserOnObjectRequest{Mappings: commentMappings})
			if err != nil {
				return xerror.Wrapf(err, "failed to get comment status").WithCtx(ctx)
			}

			pairs := resp.GetResults()
			for _, comInfo := range pairs[uid].GetList() {
				oidCommented[comInfo.Oid] = comInfo.Commented
			}
			for _, note := range notes {
				noteId := int64(note.NoteId)
				note.Interact.Commented = oidCommented[noteId]
			}
			return nil
		})
	})

	if err := eg.Wait(); err != nil {
		xlog.Msgf("failed to assign note extra").Err(err).Errorx(ctx)
		return
	}

	for _, note := range notes {
		noteId := int64(note.NoteId)
		note.Interact.Liked = oidLiked[noteId]
		note.Interact.Commented = oidCommented[noteId]
	}
}

// LikeNote 点赞/取消点赞笔记
func (b *Biz) LikeNote(ctx context.Context, req *model.LikeNoteReq) error {
	uid := metadata.Uid(ctx)
	_, err := dep.NoteInteractServer().LikeNote(ctx, &notev1.LikeNoteRequest{
		NoteId:    int64(req.NoteId),
		Uid:       uid,
		Operation: notev1.LikeNoteRequest_Operation(req.Action),
	})
	return err
}

// GetNoteLikeCount 获取笔记点赞数
func (b *Biz) GetNoteLikeCount(ctx context.Context, noteId imodel.NoteId) (*model.GetNoteLikeCountRes, error) {
	resp, err := dep.NoteInteractServer().GetNoteLikes(ctx,
		&notev1.GetNoteLikesRequest{NoteId: int64(noteId)})
	if err != nil {
		return nil, err
	}

	return &model.GetNoteLikeCountRes{
		NoteId: imodel.NoteId(resp.NoteId),
		Count:  resp.Likes,
	}, nil
}
