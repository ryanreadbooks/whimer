package biz

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/feed/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/feed/internal/model"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	maps "github.com/ryanreadbooks/whimer/misc/xmap"
	slices "github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"

	"golang.org/x/sync/errgroup"
)

type FeedBiz interface {
	RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) ([]*model.FeedNoteItem, error)
	GetNote(ctx context.Context, noteId uint64) (*model.FeedNoteItem, error)
}

type feedBiz struct {
}

func NewFeedBiz() FeedBiz { return &feedBiz{} }

// 收集作者信息
func (b *feedBiz) collectAuthor(ctx context.Context, uids []int64) (map[int64]*userv1.UserInfo, error) {
	authors := make(map[int64]*userv1.UserInfo)
	resp, err := dep.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
		Uids: uids,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to get user infos").WithExtra("uids", uids).WithCtx(ctx)
	}

	for _, u := range resp.GetUsers() {
		authors[u.Uid] = u
	}

	return authors, nil
}

// 收集reqUid和noteIds之间的评论关系
func (b *feedBiz) collectCommentStatus(ctx context.Context, reqUid int64, noteIds []uint64) (map[uint64]bool, error) {
	oidCommented := make(map[uint64]bool)
	// uid -> [oid1, oid2, ...]
	objs := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
	objs[reqUid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{
		Oids: noteIds,
	}
	resp, err := dep.Commenter().BatchCheckUserOnObject(ctx,
		&commentv1.BatchCheckUserOnObjectRequest{
			Mappings: objs,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to check user on object").WithCtx(ctx)
	}

	// organize result
	pairs := resp.GetResults()
	for _, comInfo := range pairs[reqUid].GetList() {
		oidCommented[comInfo.Oid] = comInfo.Commented
	}

	return oidCommented, nil
}

// 获取评论数量
func (b *feedBiz) collectCommentNumber(ctx context.Context, noteIds []uint64) (map[uint64]uint64, error) {
	resp, err := dep.Commenter().BatchCountReply(ctx, &commentv1.BatchCountReplyRequest{
		Oids: noteIds,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to count reply").WithCtx(ctx)
	}

	commentNums := resp.GetNumbers()

	return commentNums, nil
}

// 获取reqUid和noteIds之间的点赞关系
func (b *feedBiz) collectLikeStatus(ctx context.Context, reqUid int64, noteIds []uint64) (map[uint64]bool, error) {
	oidLiked := make(map[uint64]bool)
	mappings := make(map[int64]*notev1.NoteIdList)
	mappings[reqUid] = &notev1.NoteIdList{NoteIds: noteIds}
	resp, err := dep.NoteInteracter().BatchCheckUserLikeStatus(ctx, &notev1.BatchCheckUserLikeStatusRequest{
		Mappings: mappings,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to batch check user like status").WithCtx(ctx)
	}

	pairs := resp.GetResults()
	for _, likedInfo := range pairs[reqUid].GetList() {
		oidLiked[likedInfo.NoteId] = likedInfo.Liked
	}

	return oidLiked, nil
}

func (b *feedBiz) collectRelationStatus(ctx context.Context, reqUid int64, authorUids []int64) (map[int64]bool, error) {
	resp, err := dep.Relationer().BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     reqUid,
			Targets: authorUids,
		})
	if err != nil {
		return nil, xerror.Wrapf(err,
			"feed biz failed to batch check user following authors status").WithCtx(ctx)
	}

	return resp.Status, nil
}

func (b *feedBiz) assembleNoteFeedReturn(ctx context.Context, notes []*notev1.FeedNoteItem) (
	[]*model.FeedNoteItem, error) {
	var (
		err     error
		reqUid  = metadata.Uid(ctx)
		authors = make(map[int64][]uint64, len(notes)) // 作者，一个作者可能对应多篇笔记
	)

	for _, note := range notes {
		authors[note.Author] = append(authors[note.Author], note.NoteId)
	}

	var (
		authorInfos  map[int64]*userv1.UserInfo // uid -> author info
		oidCommented map[uint64]bool            // oid -> reqUid commented or not
		oidLiked     map[uint64]bool            // oid -> reqUid liked or not
		commentNums  map[uint64]uint64          // oid -> comment count
		userFollows  map[int64]bool             // authorId -> isFollowed
	)

	noteIds := make([]uint64, 0, len(notes)) // 全部笔记id
	for _, n := range notes {
		noteIds = append(noteIds, n.NoteId)
	}

	eg, ctx := errgroup.WithContext(ctx)
	authorUids := slices.Uniq(maps.Keys(authors))
	// 2. 获取各篇笔记的作者信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			uids := authorUids
			var err error
			authorInfos, err = b.collectAuthor(ctx, uids)
			return err
		})
	})

	// 3. 获取各篇笔记当前用户的交互信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidCommented, err = b.collectCommentStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 4. 获取评论数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			commentNums, err = b.collectCommentNumber(ctx, noteIds)
			return err
		})
	})

	// 5. 获取reqUid对笔记的点赞
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			oidLiked, err = b.collectLikeStatus(ctx, reqUid, noteIds)
			return err
		})
	})

	// 6. 获取reqUid对笔记作者的关注状态
	eg.Go(func() error {
		return recovery.Do(func() error {
			var err error
			userFollows, err = b.collectRelationStatus(ctx, reqUid, authorUids)
			if err != nil {
				xlog.Msg("feed biz failed to collect relation status").Extras("authors", authors).Err(err).Errorx(ctx)
			}

			// 非关键数据降级处理
			if userFollows == nil {
				userFollows = make(map[int64]bool)
			}

			return nil
		})
	})

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	// organize final results
	feedNotes := make([]*model.FeedNoteItem, 0, len(notes))
	for _, note := range notes {
		author := authorInfos[note.Author]
		if author == nil {
			continue
		}

		feedNote := model.NewFeedNoteItemFromPb(note)
		noteId := feedNote.NoteId
		// fill extra fields
		feedNote.Author = model.NewAuthor(author)
		feedNote.Comments = commentNums[noteId]
		feedNote.Interact.Commented = oidCommented[noteId]
		feedNote.Interact.Liked = oidLiked[noteId]
		feedNote.Interact.Followed = userFollows[author.Uid]

		feedNotes = append(feedNotes, feedNote)
	}

	return feedNotes, nil
}

func (b *feedBiz) RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) ([]*model.FeedNoteItem, error) {
	// 1. 获取笔记基础信息
	resp, err := dep.NoteFeeder().RandomGet(ctx, &notev1.RandomGetRequest{
		Count: int32(req.NeedNum),
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz random get note failed").WithExtras("req", req).WithCtx(ctx)
	}

	notes := resp.GetItems()
	if len(notes) == 0 {
		return []*model.FeedNoteItem{}, nil
	}

	// 2. 组装所有需要的信息
	return b.assembleNoteFeedReturn(ctx, notes)
}

func (b *feedBiz) GetNote(ctx context.Context, noteId uint64) (*model.FeedNoteItem, error) {
	// 1. 获取指定笔记
	resp, err := dep.NoteFeeder().GetFeedNote(ctx, &notev1.GetFeedNoteRequest{
		NoteId: noteId,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz failed to get note").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	note := resp.GetItem()
	feeds, err := b.assembleNoteFeedReturn(ctx, []*notev1.FeedNoteItem{note})
	if err != nil {
		return nil, err
	}

	return feeds[0], nil
}
