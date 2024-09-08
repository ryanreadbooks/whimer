package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/counter/internal/svc"
	"github.com/ryanreadbooks/whimer/counter/sdk/v1"

	"github.com/bufbuild/protovalidate-go"
)

type CounterServer struct {
	v1.UnimplementedCounterServiceServer
	validator *protovalidate.Validator

	Svc *svc.ServiceContext
}

func NewCounterServer(ctx *svc.ServiceContext) *CounterServer {
	validator, err := protovalidate.New()
	if err != nil {
		panic(err)
	}
	return &CounterServer{
		Svc:       ctx,
		validator: validator,
	}
}

func (s *CounterServer) AddRecord(ctx context.Context, req *v1.AddRecordRequest) (*v1.AddRecordResponse, error) {
	return s.Svc.RecordSvc.AddRecord(ctx, req)
}

func (s *CounterServer) CancelRecord(ctx context.Context, req *v1.CancelRecordRequest) (*v1.CancelRecordResponse, error) {
	return s.Svc.RecordSvc.CancelRecord(ctx, req)
}

func (s *CounterServer) GetRecord(ctx context.Context, req *v1.GetRecordRequest) (*v1.GetRecordResponse, error) {
	return s.Svc.RecordSvc.GetRecord(ctx, req)
}

func (s *CounterServer) GetSummary(ctx context.Context, req *v1.GetSummaryRequest) (*v1.GetSummaryResponse, error) {
	return s.Svc.RecordSvc.GetSummary(ctx, req)
}