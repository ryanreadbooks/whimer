package comment

import (
	"context"
	"sync"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/dto"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	commentvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
	userdomain "github.com/ryanreadbooks/whimer/pilot/internal/domain/user"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
)

type Service struct {
	commentAdapter            repository.CommentAdapter
	userAdapter               userrepo.UserServiceAdapter
	userDomainService         *userdomain.DomainService
	noteFeedAdapter           noterepo.NoteFeedAdapter
	noteCreatorAdapter        noterepo.NoteCreatorAdapter
	storageAdapter            storage.Repository
	systemNotifyDomainService *systemnotify.DomainService
}

func NewService(
	commentAdapter repository.CommentAdapter,
	userDomainService *userdomain.DomainService,
	userAdapter userrepo.UserServiceAdapter,
	noteFeedAdapter noterepo.NoteFeedAdapter,
	noteCreatorAdapter noterepo.NoteCreatorAdapter,
	storageAdapter storage.Repository,
	systemNotifyDomainService *systemnotify.DomainService,
) *Service {
	return &Service{
		commentAdapter:            commentAdapter,
		userDomainService:         userDomainService,
		userAdapter:               userAdapter,
		noteFeedAdapter:           noteFeedAdapter,
		noteCreatorAdapter:        noteCreatorAdapter,
		storageAdapter:            storageAdapter,
		systemNotifyDomainService: systemNotifyDomainService,
	}
}

func (s *Service) checkNoteExist(ctx context.Context, noteId int64) error {
	authorUid, err := s.noteFeedAdapter.GetNoteAuthorUid(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "get note author uid failed").WithCtx(ctx)
	}

	// 私有的笔记不能非作者评论
	uid := metadata.Uid(ctx)
	if uid != authorUid {
		_, _, err = s.noteFeedAdapter.GetNote(ctx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "get note failed").WithCtx(ctx)
		}
	} else {
		_, err = s.noteCreatorAdapter.GetNote(ctx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "get note failed for creator").WithCtx(ctx)
		}
	}

	return nil
}

// PublishComment 发布评论
func (s *Service) PublishComment(ctx context.Context, cmd *dto.PublishCommentCommand) (int64, error) {
	// 校验回复的用户是否正确
	if err := s.checkPubReqValid(ctx, cmd); err != nil {
		return 0, err
	}

	// 检查笔记是否存在
	if err := s.checkNoteExist(ctx, int64(cmd.Oid)); err != nil {
		return 0, err
	}

	commentId, err := s.commentAdapter.AddComment(ctx, cmd.ToRepoParams())
	if err != nil {
		return 0, xerror.Wrapf(err, "comment adapter add comment failed").WithCtx(ctx)
	}

	s.AfterCommentPublished(ctx, commentId, cmd)

	return commentId, nil
}

func (s *Service) checkPubReqValid(ctx context.Context, cmd *dto.PublishCommentCommand) error {
	if cmd.PubOnOidDirectly() {
		// 直接在笔记上评论，校验笔记作者
		author, err := s.noteFeedAdapter.GetNoteAuthorUid(ctx, int64(cmd.Oid))
		if err != nil {
			return xerror.Wrapf(err, "get note author failed").WithCtx(ctx)
		}
		if author != cmd.ReplyUid {
			return xerror.Wrap(errors.ErrReplyUserDoesNotMatch)
		}
	} else {
		// 在评论上回复，校验父评论的发布者
		uid, err := s.commentAdapter.GetCommentUser(ctx, cmd.ParentId)
		if err != nil {
			return xerror.Wrapf(err, "get comment user failed").WithCtx(ctx)
		}
		if uid != cmd.ReplyUid {
			return xerror.Wrap(errors.ErrReplyUserDoesNotMatch)
		}
	}

	return nil
}

// PageGetRootComments 分页获取主评论
func (s *Service) PageGetRootComments(ctx context.Context, q *dto.GetCommentsQuery) (*dto.CommentListResult, error) {
	// 检查笔记是否存在
	if err := s.checkNoteExist(ctx, int64(q.Oid)); err != nil {
		return nil, err
	}

	result, err := s.commentAdapter.PageGetRootComments(ctx, &repository.PageGetCommentsParams{
		Oid:    int64(q.Oid),
		Cursor: q.Cursor,
		SortBy: commentvo.SortType(q.SortBy),
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter page get root comments failed").WithCtx(ctx)
	}

	items := make([]*dto.Comment, 0, len(result.Items))
	if len(result.Items) > 0 {
		items, err = s.attachCommentsUsers(ctx, result.Items)
		if err != nil {
			return nil, xerror.Wrapf(err, "attach comments users failed").WithCtx(ctx)
		}
		s.attachCommentWithInteract(ctx, items)
	}

	return &dto.CommentListResult{
		Items:      items,
		NextCursor: result.NextCursor,
		HasNext:    result.HasNext,
	}, nil
}

// PageGetSubComments 分页获取子评论
func (s *Service) PageGetSubComments(ctx context.Context, q *dto.GetSubCommentsQuery) (*dto.CommentListResult, error) {
	// 检查笔记是否存在
	if err := s.checkNoteExist(ctx, int64(q.Oid)); err != nil {
		return nil, err
	}

	result, err := s.commentAdapter.PageGetSubComments(ctx, &repository.PageGetSubCommentsParams{
		Oid:    int64(q.Oid),
		RootId: q.RootId,
		Cursor: q.Cursor,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter page get sub comments failed").WithCtx(ctx)
	}

	items := make([]*dto.Comment, 0, len(result.Items))
	if len(result.Items) > 0 {
		items, err = s.attachCommentsUsers(ctx, result.Items)
		if err != nil {
			return nil, xerror.Wrapf(err, "attach comments users failed").WithCtx(ctx)
		}
		s.attachCommentWithInteract(ctx, items)
	}

	return &dto.CommentListResult{
		Items:      items,
		NextCursor: result.NextCursor,
		HasNext:    result.HasNext,
	}, nil
}

// PageGetComments 获取主评论信息（包含其下子评论）
func (s *Service) PageGetComments(ctx context.Context, q *dto.GetCommentsQuery) (*dto.DetailedCommentListResult, error) {
	var (
		wg sync.WaitGroup

		pinned *dto.DetailedComment
		seeked *dto.DetailedComment
	)

	oid := int64(q.Oid)

	// 检查笔记是否存在
	if err := s.checkNoteExist(ctx, oid); err != nil {
		return nil, err
	}

	if q.Cursor == 0 {
		wg.Add(1)
		// 第一次请求需要返回置顶评论
		concurrent.SafeGo(func() {
			defer wg.Done()
			var err error
			pinned, err = s.getPinnedComment(ctx, oid)
			if err != nil {
				xlog.Msg("get pinned comment failed").Extras("query", q).Err(err).Errorx(ctx)
			}
		})
	}

	// 检查是否需要获取指定评论
	if q.SeekId != 0 {
		wg.Add(1)
		concurrent.SafeGo(func() {
			defer wg.Done()
			var err error
			seeked, err = s.getSeekedComment(ctx, oid, q.SeekId)
			if err != nil {
				xlog.Msg("get seeked comment failed").Extras("query", q).Err(err).Errorx(ctx)
			}
		})
	}

	result, err := s.commentAdapter.PageGetDetailedComments(ctx,
		&repository.PageGetDetailedCommentsParams{
			Oid:    oid,
			Cursor: q.Cursor,
			SortBy: commentvo.SortType(q.SortBy),
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter page get detailed comments failed").WithCtx(ctx)
	}

	comments := make([]*dto.DetailedComment, 0, len(result.Items))
	if len(result.Items) > 0 {
		uidsMap := s.extractUidsMap(result.Items)
		userMap, err := s.userAdapter.BatchGetUser(ctx, xmap.Keys(uidsMap))
		if err != nil {
			return nil, xerror.Wrapf(err, "user adapter batch get user failed").WithCtx(ctx)
		}

		for _, item := range result.Items {
			dtoItem := dto.EntityToDetailedComment(item)
			s.fillUserInfo(ctx, dtoItem, userMap)
			comments = append(comments, dtoItem)
		}
	}

	wg.Wait()

	// 合并需要附加交互信息的评论
	temps := make([]*dto.DetailedComment, 0, len(comments)+2)
	temps = append(temps, comments...)
	if pinned != nil {
		temps = append(temps, pinned)
	}
	if q.SeekId != 0 && seeked != nil {
		temps = append(temps, seeked)
	}

	s.attachDetailedCommentWithInteract(ctx, temps)

	// 在开头加上seeked评论
	if q.SeekId != 0 && seeked != nil {
		newComments := make([]*dto.DetailedComment, 0, len(comments)+1)
		newComments = append(newComments, seeked)
		newComments = append(newComments, comments...)
		comments = xslice.UniqF(newComments, func(v *dto.DetailedComment) int64 { return v.Root.Id })
	}

	return &dto.DetailedCommentListResult{
		Comments:   comments,
		PinComment: pinned,
		NextCursor: result.NextCursor,
		HasNext:    result.HasNext,
	}, nil
}

func (s *Service) getPinnedComment(ctx context.Context, oid int64) (*dto.DetailedComment, error) {
	item, err := s.commentAdapter.GetPinnedComment(ctx, oid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get pinned comment failed").WithCtx(ctx)
	}
	if item == nil || item.Root == nil {
		return nil, nil
	}

	uidsMap := s.extractUidsMap([]*entity.DetailedComment{item})
	userMap, err := s.userAdapter.BatchGetUser(ctx, xmap.Keys(uidsMap))
	if err != nil {
		return nil, xerror.Wrapf(err, "user adapter batch get user failed").WithCtx(ctx)
	}

	dtoItem := dto.EntityToDetailedComment(item)
	s.fillUserInfo(ctx, dtoItem, userMap)

	return dtoItem, nil
}

func (s *Service) getSeekedComment(ctx context.Context, oid, seekId int64) (*dto.DetailedComment, error) {
	seekComment, err := s.commentAdapter.GetComment(ctx, seekId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get comment failed").WithCtx(ctx)
	}
	if seekComment.Oid != oid {
		return nil, nil
	}

	var seekedItem *entity.DetailedComment

	if seekComment.RootId == 0 && seekComment.ParentId == 0 {
		// seekComment是主评论，需要获取子评论
		subResult, err := s.commentAdapter.PageGetSubComments(ctx, &repository.PageGetSubCommentsParams{
			Oid:    seekComment.Oid,
			RootId: seekComment.Id,
			Cursor: 0,
		})
		if err != nil {
			return nil, xerror.Wrapf(err, "get sub comments failed").WithCtx(ctx)
		}

		seekedItem = &entity.DetailedComment{
			Root: seekComment,
			SubComments: &entity.SubComments{
				Items:      subResult.Items,
				HasNext:    subResult.HasNext,
				NextCursor: subResult.NextCursor,
			},
		}
	} else {
		// seekComment是子评论，需要获取主评论
		rootComment, err := s.commentAdapter.GetComment(ctx, seekComment.RootId)
		if err != nil {
			return nil, xerror.Wrapf(err, "get root comment failed").WithCtx(ctx)
		}

		seekedItem = &entity.DetailedComment{
			Root: rootComment,
			SubComments: &entity.SubComments{
				Items:      []*entity.Comment{seekComment},
				NextCursor: 0,
				HasNext:    true,
			},
		}
	}

	uidsMap := s.extractUidsMap([]*entity.DetailedComment{seekedItem})
	userMap, err := s.userAdapter.BatchGetUser(ctx, xmap.Keys(uidsMap))
	if err != nil {
		return nil, xerror.Wrapf(err, "user adapter batch get user failed").WithCtx(ctx)
	}

	dtoItem := dto.EntityToDetailedComment(seekedItem)
	s.fillUserInfo(ctx, dtoItem, userMap)

	return dtoItem, nil
}

// DeleteComment 删除评论
func (s *Service) DeleteComment(ctx context.Context, cmd *dto.DeleteCommentCommand) error {
	err := s.commentAdapter.DelComment(ctx, cmd.CommentId, int64(cmd.Oid))
	if err != nil {
		return xerror.Wrapf(err, "comment adapter del comment failed").WithCtx(ctx)
	}

	return nil
}

// PinComment 置顶/取消置顶评论
func (s *Service) PinComment(ctx context.Context, cmd *dto.PinCommentCommand) error {
	// 检查笔记是否存在
	if err := s.checkNoteExist(ctx, int64(cmd.Oid)); err != nil {
		return err
	}

	err := s.commentAdapter.PinComment(ctx, int64(cmd.Oid), cmd.CommentId, cmd.Action)
	if err != nil {
		return xerror.Wrapf(err, "comment adapter pin comment failed").WithCtx(ctx)
	}

	return nil
}

// LikeComment 点赞评论
func (s *Service) LikeComment(ctx context.Context, cmd *dto.LikeCommentCommand) error {
	cmt, err := s.commentAdapter.GetComment(ctx, cmd.CommentId)
	if err != nil {
		return xerror.Wrapf(err, "get comment failed").WithCtx(ctx)
	}

	err = s.commentAdapter.LikeComment(ctx, cmd.CommentId, cmd.Action)
	if err != nil {
		return xerror.Wrapf(err, "comment adapter like comment failed").WithCtx(ctx)
	}

	operator := metadata.Uid(ctx)
	if cmt.Uid == 0 || cmt.Uid == operator { // 自己给自己评论不用管
		return nil
	}

	//  通知被点赞的用户
	err = s.systemNotifyDomainService.NotifyUserLikesOnComment(ctx, operator, cmt.Uid,
		&notifyvo.NotifyLikesOnCommentParam{
			NoteId:    notevo.NoteId(cmt.Oid),
			CommentId: cmd.CommentId,
		})
	if err != nil {
		// log only
		xlog.Msg("notify user likes on comment failed").Err(err).Errorx(ctx)
	}

	return nil
}

// DislikeComment 点踩评论
func (s *Service) DislikeComment(ctx context.Context, cmd *dto.DislikeCommentCommand) error {
	err := s.commentAdapter.DislikeComment(ctx, cmd.CommentId, cmd.Action)
	if err != nil {
		return xerror.Wrapf(err, "comment adapter dislike comment failed").WithCtx(ctx)
	}

	return nil
}

// GetCommentLikeCount 获取评论点赞数
func (s *Service) GetCommentLikeCount(ctx context.Context, commentId int64) (*dto.LikeCountResult, error) {
	count, err := s.commentAdapter.GetCommentLikeCount(ctx, commentId)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter get comment like count failed").WithCtx(ctx)
	}

	return &dto.LikeCountResult{
		CommentId: commentId,
		Likes:     count,
	}, nil
}

// GetComment 获取评论详情
func (s *Service) GetComment(ctx context.Context, commentId int64) (*dto.Comment, error) {
	item, err := s.commentAdapter.GetComment(ctx, commentId)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter get comment failed").WithCtx(ctx)
	}

	return dto.EntityToComment(item), nil
}

// GetCommentContent 获取评论内容
func (s *Service) GetCommentContent(ctx context.Context, commentId int64) (*dto.CommentContent, error) {
	item, err := s.commentAdapter.GetComment(ctx, commentId)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment adapter get comment failed").WithCtx(ctx)
	}

	atUsers := make([]commondto.AtUser, 0, len(item.AtUsers))
	for _, au := range item.AtUsers {
		atUsers = append(atUsers, commondto.AtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	return &dto.CommentContent{
		Text:    item.Content,
		AtUsers: atUsers,
	}, nil
}

// 附加用户信息
func (s *Service) attachCommentsUsers(ctx context.Context, items []*entity.Comment) ([]*dto.Comment, error) {
	uidsMap := make(map[int64]struct{})
	for _, item := range items {
		uidsMap[item.Uid] = struct{}{}
	}

	userMap, err := s.userAdapter.BatchGetUser(ctx, xmap.Keys(uidsMap))
	if err != nil {
		return nil, xerror.Wrapf(err, "user adapter batch get user failed").WithCtx(ctx)
	}

	result := make([]*dto.Comment, 0, len(items))
	for _, item := range items {
		dtoItem := dto.EntityToComment(item)
		dtoItem.User = dto.UserVoToCommentUser(userMap[item.Uid])
		dtoItem.IpLoc, _ = infra.Ip2Loc().Convert(ctx, dtoItem.Ip)
		result = append(result, dtoItem)
	}

	return result, nil
}

// 附加评论交互信息
func (s *Service) attachCommentWithInteract(ctx context.Context, items []*dto.Comment) {
	uid := metadata.Uid(ctx)
	if uid == 0 || len(items) == 0 {
		return
	}
	commentIds := make([]int64, 0, len(items))
	for _, item := range items {
		commentIds = append(commentIds, item.Id)
	}
	commentIds = xslice.Uniq(commentIds)

	likeStatus, err := s.commentAdapter.BatchCheckUserLikeComment(ctx, uid, commentIds)
	if err != nil {
		xlog.Msg("batch check user like comment failed").Err(err).Errorx(ctx)
		return
	}

	for _, item := range items {
		if liked, ok := likeStatus[item.Id]; ok {
			item.Interact.Liked = liked
		}
	}
}

func (s *Service) attachDetailedCommentWithInteract(ctx context.Context, dItems []*dto.DetailedComment) {
	items := make([]*dto.Comment, 0, len(dItems))
	for _, dItem := range dItems {
		items = append(items, dItem.Root)
		items = append(items, dItem.SubComments.Items...)
	}

	s.attachCommentWithInteract(ctx, items)
}

func (s *Service) extractUidsMap(items []*entity.DetailedComment) map[int64]struct{} {
	uidsMap := make(map[int64]struct{})
	for _, item := range items {
		uidsMap[item.Root.Uid] = struct{}{}
		for _, sub := range item.SubComments.Items {
			uidsMap[sub.Uid] = struct{}{}
		}
	}
	return uidsMap
}

func (s *Service) fillUserInfo(ctx context.Context, dtoItem *dto.DetailedComment, userMap map[int64]*uservo.User) {
	if dtoItem.Root != nil {
		dtoItem.Root.User = dto.UserVoToCommentUser(userMap[dtoItem.Root.Uid])
		dtoItem.Root.IpLoc, _ = infra.Ip2Loc().Convert(ctx, dtoItem.Root.Ip)
	}
	for _, sub := range dtoItem.SubComments.Items {
		sub.User = dto.UserVoToCommentUser(userMap[sub.Uid])
		sub.IpLoc, _ = infra.Ip2Loc().Convert(ctx, sub.Ip)
	}
}

func (s *Service) GetUploadImageTicket(ctx context.Context, count int32) (*storagevo.UploadTicketDeprecated, error) {
	tickets, err := s.storageAdapter.GetUploadTicketDeprecated(ctx,
		storagevo.ObjectTypeCommentImage, count)
	if err != nil {
		return nil, xerror.Wrapf(err, "storage adapter get upload image ticket failed").WithCtx(ctx)
	}

	return tickets, nil
}
