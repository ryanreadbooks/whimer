package grpc

import (
	"context"

	namespacev1 "github.com/ryanreadbooks/whimer/conductor/api/namespace/v1"
	namespaceservice "github.com/ryanreadbooks/whimer/conductor/api/namespaceservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
)

type NamespaceServiceServer struct {
	namespaceservice.UnimplementedNamespaceServiceServer

	srv *service.Service
}

func NewNamespaceServiceServer(srv *service.Service) *NamespaceServiceServer {
	return &NamespaceServiceServer{
		srv: srv,
	}
}

// CreateNamespace 创建命名空间
func (s *NamespaceServiceServer) CreateNamespace(ctx context.Context,
	in *namespaceservice.CreateNamespaceRequest) (*namespaceservice.CreateNamespaceResponse, error) {
	ns, err := s.srv.NamespaceService.CreateNamespace(ctx, in.Name)
	if err != nil {
		return nil, err
	}

	return &namespaceservice.CreateNamespaceResponse{
		Namespace: &namespacev1.Namespace{
			Id:   ns.Id,
			Name: ns.Name,
		},
	}, nil
}

// ListNamespace 分页列出命名空间
func (s *NamespaceServiceServer) ListNamespace(ctx context.Context,
	in *namespaceservice.ListNamespaceRequest) (*namespaceservice.ListNamespaceResponse, error) {
	result, err := s.srv.NamespaceService.ListNamespace(ctx, int(in.Page), int(in.Count))
	if err != nil {
		return nil, err
	}

	namespaces := make([]*namespacev1.Namespace, 0, len(result.Namespaces))
	for _, ns := range result.Namespaces {
		namespaces = append(namespaces, &namespacev1.Namespace{
			Id:   ns.Id,
			Name: ns.Name,
		})
	}

	return &namespaceservice.ListNamespaceResponse{
		Namespaces: namespaces,
		Total:      result.Total,
	}, nil
}
