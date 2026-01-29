package repo

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/recentcontact"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/repo/user"
)

var recentContactRepo *user.RecentContactRepo

func Init(recentContactStore *recentcontact.Store) {
	recentContactRepo = user.NewRecentContactRepo(recentContactStore)
}

func RecentContactRepo() *user.RecentContactRepo {
	return recentContactRepo
}
