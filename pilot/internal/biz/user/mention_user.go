package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"

	"golang.org/x/sync/errgroup"
)

func (b *Biz) GetMentionUserCandidates(ctx context.Context, uid int64, search string) ([]*model.MentionUserRespItem, error) {
	// 分组获取@用户
	eg, ctx := errgroup.WithContext(ctx)

	groups := make([]*model.MentionUserRespItem, 3)
	// 拿最近联系人
	eg.Go(recovery.DoV2(func() error {
		recentContacts, err := b.GetAllRecentContacts(ctx, uid)
		if err != nil {
			xlog.Msg("get recent contacts failed").Err(err).Errorx(ctx)
		}

		groups[0] = &model.MentionUserRespItem{
			Group:     model.MentionRecentContacts,
			GroupDesc: model.MentionRecentContacts.Desc(),
			Users:     recentContacts,
		}

		return nil
	}))

	// 我的关注
	eg.Go(recovery.DoV2(func() error {
		myFollowings, err := b.BrutalListFollowingsByName(ctx, uid, search)
		if err != nil {
			xlog.Msg("list followings groups failed").Err(err).Errorx(ctx)
		}

		groups[1] = &model.MentionUserRespItem{
			Group:     model.MentionFollowings,
			GroupDesc: model.MentionFollowings.Desc(),
			Users:     myFollowings,
		}

		return nil
	}))

	// TODO 其他人 try to use elastic search in the future
	if len(search) > 0 {
		eg.Go(recovery.DoV2(func() error {
			groups[2] = &model.MentionUserRespItem{
				Group:     model.MentionOthers,
				GroupDesc: model.MentionOthers.Desc(),
				Users:     []*model.User{},
			}
			return nil
		}))
	}

	if err := eg.Wait(); err != nil {
		return nil, xerror.Wrapf(err, "get mention user candidates failed").WithCtx(ctx)
	}

	return groups, nil
}
