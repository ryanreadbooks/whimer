package rpc

import (
	"context"

	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/svc"
)

type CounterServer struct {
	counterv1.UnimplementedCounterServiceServer

	Svc *svc.ServiceContext
}

func NewCounterServer(ctx *svc.ServiceContext) *CounterServer {
	return &CounterServer{
		Svc: ctx,
	}
}

func (s *CounterServer) AddRecord(ctx context.Context, req *counterv1.AddRecordRequest) (
	*counterv1.AddRecordResponse, error) {
	return s.Svc.RecordSvc.AddRecord(ctx, req)
}

func (s *CounterServer) CancelRecord(ctx context.Context, req *counterv1.CancelRecordRequest) (
	*counterv1.CancelRecordResponse, error) {
	return s.Svc.RecordSvc.CancelRecord(ctx, req)
}

func (s *CounterServer) GetRecord(ctx context.Context, req *counterv1.GetRecordRequest) (
	*counterv1.GetRecordResponse, error) {
	return s.Svc.RecordSvc.GetRecord(ctx, req)
}

func (s *CounterServer) BatchGetRecord(ctx context.Context, req *counterv1.BatchGetRecordRequest) (
	*counterv1.BatchGetRecordResponse, error) {
	var uidOids = make(map[int64][]int64, len(req.Params))
	for uid, oids := range req.Params {
		uidOids[uid] = append(uidOids[uid], oids.Oids...)
	}

	resp, err := s.Svc.RecordSvc.BatchGetRecord(ctx, uidOids, int(req.BizCode))
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
	return s.Svc.RecordSvc.GetSummary(ctx, req)
}

func (s *CounterServer) BatchGetSummary(ctx context.Context, req *counterv1.BatchGetSummaryRequest) (
	*counterv1.BatchGetSummaryResponse, error) {
	return s.Svc.RecordSvc.BatchGetSummary(ctx, req)
}
