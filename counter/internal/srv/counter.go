package srv

import (
	"context"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/biz"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type CounterSrv struct {
	CounterBiz *biz.CounterBiz
}

func NewCounterSrv(s *Service, biz *biz.Biz) *CounterSrv {
	return &CounterSrv{
		CounterBiz: biz.CounterBiz,
	}
}

func (s *CounterSrv) AddRecord(ctx context.Context, req *counterv1.AddRecordRequest) (*counterv1.AddRecordResponse, error) {
	return s.CounterBiz.AddRecord(ctx, req)
}

func (s *CounterSrv) CancelRecord(ctx context.Context,
	req *counterv1.CancelRecordRequest) (*counterv1.CancelRecordResponse, error) {
	return s.CounterBiz.CancelRecord(ctx, req)
}

func (s *CounterSrv) GetRecord(ctx context.Context,
	req *counterv1.GetRecordRequest) (*counterv1.GetRecordResponse, error) {
	return s.CounterBiz.GetRecord(ctx, req)
}

func (s *CounterSrv) BatchGetRecord(ctx context.Context, uidOids map[int64][]int64, biz int) (
	map[int64][]*counterv1.Record, error) {
	return s.CounterBiz.BatchGetRecord(ctx, uidOids, biz)
}

func (s *CounterSrv) GetSummary(ctx context.Context, req *counterv1.GetSummaryRequest) (
	*counterv1.GetSummaryResponse, error) {
	return s.CounterBiz.GetSummary(ctx, req)
}

func (s *CounterSrv) BatchGetSummary(ctx context.Context, req *counterv1.BatchGetSummaryRequest) (
	*counterv1.BatchGetSummaryResponse, error) {

	return s.CounterBiz.BatchGetSummary(ctx, req)
}

func (s *CounterSrv) PageListUserRecords(ctx context.Context, req *counterv1.PageGetUserRecordRequest) (
	*counterv1.PageGetUserRecordResponse, error) {

	order := biz.PageListDescOrder
	if req.SortRule == counterv1.SortRule_SORT_RULE_ASC {
		order = biz.PageListAscOrder
	}
	records, nextReq, err := s.CounterBiz.PageListRecords(ctx, req.BizCode, req.Uid,
		biz.PageListRecordsParam{
			Cursor: req.Cursor,
			Count:  req.Count,
			Order:  order,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "counter srv failed to page list records").WithCtx(ctx)
	}

	return &counterv1.PageGetUserRecordResponse{
		Items:      records,
		NextCursor: nextReq.NextCursor,
		HasNext:    nextReq.HasNext,
	}, nil
}
