package biz

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/feed/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/feed/internal/model"
	"github.com/ryanreadbooks/whimer/misc/utils/maps"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
	"golang.org/x/sync/errgroup"
)

type FeedBiz interface {
	RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) error
}

type feedBiz struct {
}

func NewFeedBiz() FeedBiz { return &feedBiz{} }

func (b *feedBiz) RandomFeed(ctx context.Context, req *model.FeedRecommendRequest) error {
	// 1. 获取笔记基础信息
	resp, err := dep.NoteFeeder().RandomGet(ctx, &notev1.RandomGetRequest{
		Count: int32(req.NeedNum),
	})
	if err != nil {
		return xerror.Wrapf(err, "feed biz random get failed").WithExtras("req", req).WithCtx(ctx)
	}

	authors := make(map[uint64][]uint64, len(resp.GetItems()))
	for _, note := range resp.GetItems() {
		authors[note.Author] = append(authors[note.Author], note.NoteId)
	}

	var (
		eg        errgroup.Group
		userInfos *userv1.BatchGetUserResponse
	)

	// 2. 获取各篇笔记的作者信息
	eg.Go(func() error {
		var err2 error
		uids := slices.Uniq(maps.Keys(authors))
		userInfos, err2 = dep.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
			Uids: uids,
		})
		if err2 != nil {
			return xerror.Wrapf(err2, "feed biz failed to get user infos").WithExtra("uids", uids).WithCtx(ctx)
		}
		return nil
	})

	// 3. 获取各篇笔记当前用户的交互信息
	eg.Go(func() error {
		dep.Commenter().CheckUserCommentOnObject(ctx, 
			&commentv1.CheckUserCommentOnObjectRequest{
			})
		return nil
	})

	err = eg.Wait()
	if err != nil {
		return err
	}

	return nil
}
