package biz

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/feed/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/feed/internal/model"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
	"golang.org/x/sync/errgroup"
)

type FeedBiz interface {
	RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) ([]*model.FeedNoteItem, error)
}

type feedBiz struct {
}

func NewFeedBiz() FeedBiz { return &feedBiz{} }

func (b *feedBiz) RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) ([]*model.FeedNoteItem, error) {
	var (
		reqUid = metadata.Uid(ctx)
	)

	// 1. 获取笔记基础信息
	resp, err := dep.NoteFeeder().RandomGet(ctx, &notev1.RandomGetRequest{
		Count: int32(req.NeedNum),
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "feed biz random get note failed").WithExtras("req", req).WithCtx(ctx)
	}

	notes := resp.GetItems()
	authors := make(map[uint64][]uint64, len(notes))
	for _, note := range resp.GetItems() {
		authors[note.Author] = append(authors[note.Author], note.NoteId)
	}

	var (
		eg           errgroup.Group
		authorInfos  = make(map[uint64]*userv1.UserInfo) // uid -> author info
		oidCommented = make(map[uint64]bool)             // oid -> reqUid commented or not
		oidLiked     = make(map[uint64]bool)             // oid -> reqUid liked or not
		commentNums  = make(map[uint64]uint64)           // oid -> comment count
	)

	// 2. 获取各篇笔记的作者信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			uids := slices.Uniq(maps.Keys(authors))
			resp, err := dep.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
				Uids: uids,
			})
			if err != nil {
				return xerror.Wrapf(err, "feed biz failed to get user infos").WithExtra("uids", uids).WithCtx(ctx)
			}

			for _, u := range resp.GetUsers() {
				authorInfos[u.Uid] = u
			}

			return nil
		})
	})

	oids := make([]uint64, 0, len(notes)) // 全部笔记id
	for _, n := range notes {
		oids = append(oids, n.NoteId)
	}

	// 3. 获取各篇笔记当前用户的交互信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			// uid -> [oid1, oid2, ...]
			objs := make(map[uint64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
			objs[reqUid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{
				Oids: oids,
			}
			resp, err := dep.Commenter().BatchCheckUserOnObject(ctx,
				&commentv1.BatchCheckUserOnObjectRequest{
					Mappings: objs,
				})
			if err != nil {
				return xerror.Wrapf(err, "feed biz failed to check user on object").WithCtx(ctx)
			}

			// organize result
			pairs := resp.GetResults()
			for _, comInfo := range pairs[reqUid].GetList() {
				oidCommented[comInfo.Oid] = comInfo.Commented
			}

			return nil
		})
	})

	// 4. 获取评论数量
	eg.Go(func() error {
		resp, err := dep.Commenter().BatchCountReply(ctx, &commentv1.BatchCountReplyRequest{
			Oids: oids,
		})
		if err != nil {
			return xerror.Wrapf(err, "feed biz failed to count reply").WithCtx(ctx)
		}

		commentNums = resp.GetNumbers()

		return nil
	})

	// 5. 获取reqUid对笔记的点赞
	eg.Go(func() error {
		mappings := make(map[uint64]*notev1.NoteIdList)
		mappings[reqUid] = &notev1.NoteIdList{NoteIds: oids}
		resp, err := dep.NoteInteracter().BatchCheckUserLikeStatus(ctx, &notev1.BatchCheckUserLikeStatusRequest{
			Mappings: mappings,
		})
		if err != nil {
			return xerror.Wrapf(err, "feed biz failed to batch check user like status").WithCtx(ctx)
		}

		pairs := resp.GetResults()
		for _, likedInfo := range pairs[reqUid].GetList() {
			oidLiked[likedInfo.NoteId] = likedInfo.Liked
		}

		return nil
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

		feedNotes = append(feedNotes, feedNote)
	}

	return feedNotes, nil
}
