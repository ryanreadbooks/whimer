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

func (s *CounterSrv) BatchGetRecord(ctx context.Context, uidOids map[int64][]int64, biz int32) (
	map[int64][]*counterv1.Record, error) {
	return s.CounterBiz.BatchGetRecord(ctx, uidOids, biz)
}

func (s *CounterSrv) GetSummary(ctx context.Context, req *counterv1.GetSummaryRequest) (
	*counterv1.GetSummaryResponse, error) {
	count, err := s.CounterBiz.GetSummary(ctx, req.BizCode, req.Oid)
	if err != nil {
		return nil, err
	}

	return &counterv1.GetSummaryResponse{
		BizCode: req.BizCode,
		Oid:     req.Oid,
		Count:   count,
	}, nil
}

func (s *CounterSrv) BatchGetSummary(ctx context.Context, req *counterv1.BatchGetSummaryRequest) (
	*counterv1.BatchGetSummaryResponse, error) {

	keys := make([]biz.SummaryKey, 0, len(req.Requests))
	for _, r := range req.Requests {
		keys = append(keys, biz.SummaryKey{BizCode: r.BizCode, Oid: r.Oid})
	}
	resp, err := s.CounterBiz.BatchGetSummary(ctx, keys)
	if err != nil {
		return nil, err
	}

	responses := make([]*counterv1.GetSummaryResponse, 0, len(resp))
	for k, v := range resp {
		responses = append(responses, &counterv1.GetSummaryResponse{
			BizCode: k.BizCode,
			Oid:     k.Oid,
			Count:   v,
		})
	}

	return &counterv1.BatchGetSummaryResponse{Responses: responses}, nil
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

func (s *CounterSrv) CheckHasActDo(ctx context.Context, req *counterv1.CheckHasActDoRequest) (bool, error) {
	return s.CounterBiz.CheckHasActDo(ctx, req)
}

func (s *CounterSrv) BatchCheckHasActDo(ctx context.Context, uidOids map[int64][]int64, biz int32) (
	map[int64][]*counterv1.BatchCheckHasActDoResponse_Item, error,
) {
	return s.CounterBiz.BatchCheckHasActDo(ctx, uidOids, biz)
}
