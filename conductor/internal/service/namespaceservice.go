package service

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
)

type NamespaceService struct {
	namespaceBiz *biz.NamespaceBiz
}

func NewNamespaceService(bizz *biz.Biz) *NamespaceService {
	return &NamespaceService{
		namespaceBiz: bizz.NamespaceBiz,
	}
}

// CreateNamespace 创建命名空间
func (s *NamespaceService) CreateNamespace(ctx context.Context, name string) (*model.Namespace, error) {
	return s.namespaceBiz.Create(ctx, name)
}

// ListNamespace 分页列出命名空间
func (s *NamespaceService) ListNamespace(ctx context.Context, page, count int) (*biz.ListNamespaceResult, error) {
	return s.namespaceBiz.List(ctx, page, count)
}
