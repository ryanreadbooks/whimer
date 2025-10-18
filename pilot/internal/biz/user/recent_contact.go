package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

// 最近联系人
//
// 用户D天内在笔记/评论中@的最多N个

// 新增最近联系人历史
func (b *Biz) AppendRecentContacts(ctx context.Context, uid int64, targets []int64) error {
	if err := b.recentContact.AtomicAppend(ctx, uid, targets); err != nil {
		return xerror.Wrapf(err, "append recent contacts failed").WithCtx(ctx)
	}

	return nil
}

func (b *Biz) AsyncAppendRecentContactsAtUser(ctx context.Context, uid int64, atUsers imodel.AtUserList) {
	targets := make([]int64, 0, len(atUsers))
	for _, atUser := range atUsers {
		if atUser.Uid == uid {
			continue
		}

		targets = append(targets, atUser.Uid)
	}

	if len(targets) <= 0 {
		return
	}

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "userbiz.atuser.append_recent_contacts",
		Job: func(ctx context.Context) error {
			if err := b.AppendRecentContacts(ctx, uid, targets); err != nil {
				xlog.Msg("userbiz append recent contacts failed").Err(err).Extras("targets", targets).Errorx(ctx)
			}

			return nil
		},
	})
}

// 获取最近所有联系人
func (b *Biz) GetAllRecentContacts(ctx context.Context, uid int64) ([]*model.User, error) {
	recents, err := b.recentContact.GetAll(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get all recent contacts failed").WithCtx(ctx)
	}

	uids := make([]int64, 0, len(recents))
	for _, recent := range recents {
		uids = append(uids, recent.Uid)
	}

	users, err := b.ListUsersV2(ctx, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "list users failed").WithCtx(ctx)
	}

	// 保持uid的返回顺序
	orderedUsers := make([]*model.User, 0, len(recents))
	for _, recent := range recents {
		orderedUsers = append(orderedUsers, users[recent.Uid])
	}

	return orderedUsers, nil
}
