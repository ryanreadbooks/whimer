package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/recentcontact"
)

var _ repository.RecentContactRepository = (*RecentContactRepo)(nil)

type RecentContactRepo struct {
	store *recentcontact.Store
}

func NewRecentContactRepo(store *recentcontact.Store) *RecentContactRepo {
	return &RecentContactRepo{
		store: store,
	}
}

func (r *RecentContactRepo) AtomicAppend(ctx context.Context, uid int64, targets []int64) error {
	return r.store.AtomicAppend(ctx, uid, targets)
}

func (r *RecentContactRepo) GetAll(ctx context.Context, uid int64) ([]vo.RecentContact, error) {
	contacts, err := r.store.GetAll(ctx, uid)
	if err != nil {
		return nil, err
	}

	result := make([]vo.RecentContact, 0, len(contacts))
	for _, c := range contacts {
		result = append(result, vo.RecentContact{
			Uid:    c.Uid,
			TimeMs: c.TimeMs,
		})
	}

	return result, nil
}
