package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/feed/internal/model"
)

func (s *service) GetRecommendFeed(ctx context.Context, req *model.FeedRecommendRequest) (
	[]*model.FeedNoteItem, error) {
	return s.FeedBiz.RandomFeed(ctx, req)
}
