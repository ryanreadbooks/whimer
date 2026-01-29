package user

import (
	"context"

	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"

	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type DomainService struct {
	userAdapter       userrepo.UserServiceAdapter
	recentContactRepo userrepo.RecentContactRepository
}

func NewDomainService(
	userAdapter userrepo.UserServiceAdapter,
	recentContactRepo userrepo.RecentContactRepository,
) *DomainService {
	return &DomainService{
		userAdapter:       userAdapter,
		recentContactRepo: recentContactRepo,
	}
}

// 新增最近联系人历史
func (s *DomainService) AppendRecentContacts(ctx context.Context, uid int64, targets []int64) error {
	return s.recentContactRepo.Append(ctx, uid, targets)
}

// at用户新增最近联系人历史
func (s *DomainService) AppendRecentContactsAtUser(ctx context.Context, uid int64, atUsers mentionvo.AtUserList) error {
	targets := make([]int64, 0, len(atUsers))
	for _, atUser := range atUsers {
		if atUser.Uid == uid {
			continue
		}
		targets = append(targets, atUser.Uid)
	}

	if len(targets) <= 0 {
		return nil
	}

	if err := s.AppendRecentContacts(ctx, uid, targets); err != nil {
		return xerror.Wrapf(err, "append recent contacts failed").WithExtras("targets", targets).WithCtx(ctx)
	}

	return nil
}

func (s *DomainService) GetAllRecentContacts(ctx context.Context, uid int64) ([]*uservo.User, error) {
	recents, err := s.recentContactRepo.GetAll(ctx, uid)
	if err != nil {
		return nil, err
	}

	uids := make([]int64, 0, len(recents))
	for _, recent := range recents {
		uids = append(uids, recent.Uid)
	}

	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := make([]*uservo.User, 0, len(recents))
	for _, recent := range recents {
		result = append(result, users[recent.Uid])
	}

	return result, nil
}
