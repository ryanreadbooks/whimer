package grpc

import (
	"context"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/srv"
)

type CounterServer struct {
	counterv1.UnimplementedCounterServiceServer

	Svc *srv.Service
}

func NewCounterServer(ctx *srv.Service) *CounterServer {
	return &CounterServer{
		Svc: ctx,
	}
}

func (s *CounterServer) AddRecord(ctx context.Context, req *counterv1.AddRecordRequest) (
	*counterv1.AddRecordResponse, error) {
	return s.Svc.CounterSrv.AddRecord(ctx, req)
}

func (s *CounterServer) CancelRecord(ctx context.Context, req *counterv1.CancelRecordRequest) (
	*counterv1.CancelRecordResponse, error) {
	return s.Svc.CounterSrv.CancelRecord(ctx, req)
}

func (s *CounterServer) GetRecord(ctx context.Context, req *counterv1.GetRecordRequest) (
	*counterv1.GetRecordResponse, error) {
	return s.Svc.CounterSrv.GetRecord(ctx, req)
}

func (s *CounterServer) BatchGetRecord(ctx context.Context, req *counterv1.BatchGetRecordRequest) (
	*counterv1.BatchGetRecordResponse, error) {
	var uidOids = make(map[int64][]int64, len(req.Params))
	for uid, oids := range req.Params {
		uidOids[uid] = append(uidOids[uid], oids.Oids...)
	}

	resp, err := s.Svc.CounterSrv.BatchGetRecord(ctx, uidOids, req.BizCode)
	if err != nil {
		return nil, err
	}

	var result = make(map[int64]*counterv1.RecordList)
	for uid, records := range resp {
		result[uid] = &counterv1.RecordList{List: records}
	}

	return &counterv1.BatchGetRecordResponse{Results: result}, nil
}

func (s *CounterServer) GetSummary(ctx context.Context, req *counterv1.GetSummaryRequest) (
	*counterv1.GetSummaryResponse, error) {
	return s.Svc.CounterSrv.GetSummary(ctx, req)
}

func (s *CounterServer) BatchGetSummary(ctx context.Context, req *counterv1.BatchGetSummaryRequest) (
	*counterv1.BatchGetSummaryResponse, error) {
	return s.Svc.CounterSrv.BatchGetSummary(ctx, req)
}

func (s *CounterServer) PageGetUserRecord(ctx context.Context, req *counterv1.PageGetUserRecordRequest) (
	*counterv1.PageGetUserRecordResponse, error) {
	return s.Svc.CounterSrv.PageListUserRecords(ctx, req)
}

// 获取一条(ActDo)计数记录
func (s *CounterServer) CheckHasActDo(ctx context.Context, in *counterv1.CheckHasActDoRequest) (
	*counterv1.CheckHasActDoResponse, error) {
	resp, err := s.Svc.CounterSrv.CheckHasActDo(ctx, in)
	if err != nil {
		return nil, err
	}

	return &counterv1.CheckHasActDoResponse{
		Do: resp,
	}, nil
}

// 批量获取(ActDo)计数记录
func (s *CounterServer) BatchCheckHasActDo(ctx context.Context, req *counterv1.BatchCheckHasActDoDoRequest) (
	*counterv1.BatchCheckHasActDoResponse, error) {
	var uidOids = make(map[int64][]int64, len(req.Params))
	for uid, oids := range req.Params {
		uidOids[uid] = append(uidOids[uid], oids.Oids...)
	}

	resp, err := s.Svc.CounterSrv.BatchCheckHasActDo(ctx, uidOids, req.BizCode)
	if err != nil {
		return nil, err
	}

	results := make(map[int64]*counterv1.BatchCheckHasActDoResponse_ItemList, len(resp))
	for uid, items := range resp {
		results[uid] = &counterv1.BatchCheckHasActDoResponse_ItemList{
			List: items,
		}
	}

	return &counterv1.BatchCheckHasActDoResponse{
		Results: results,
	}, nil
}
