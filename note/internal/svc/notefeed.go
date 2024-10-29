package svc

import (
	"context"

	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"
)

// 信息流随机获取最多count条笔记
func (s *NoteSvc) FeedRandomGet(ctx context.Context, count int32) (*notemodel.BatchNoteItem, error) {
	return s.randomGet(ctx, int(count))
}
