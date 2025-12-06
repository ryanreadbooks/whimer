package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/conductor/internal/global"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type NamespaceBiz struct {
	namespaceDao *dao.NamespaceDao
}

func NewNamespaceBiz(namespaceDao *dao.NamespaceDao) *NamespaceBiz {
	return &NamespaceBiz{
		namespaceDao: namespaceDao,
	}
}

func (b *NamespaceBiz) Create(ctx context.Context, name string) (*model.Namespace, error) {
	po := &dao.NamespacePO{
		Id:   uuid.NewUUID(),
		Name: name,
	}

	err := b.namespaceDao.Insert(ctx, po)
	if err != nil {
		if xsql.IsDuplicate(err) {
			return nil, global.ErrNamespaceAlreadyExists
		}

		return nil, xerror.Wrapf(err, "namespace biz create failed").WithExtra("name", name).WithCtx(ctx)
	}

	return model.NamespaceFromPO(po), nil
}

func (b *NamespaceBiz) Get(ctx context.Context, name string) (*model.Namespace, error) {
	po, err := b.namespaceDao.GetByName(ctx, name)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrNamespaceNotFound
		}

		return nil, xerror.Wrapf(err, "namespace dao get by name failed").WithExtra("name", name).WithCtx(ctx)
	}

	return model.NamespaceFromPO(po), nil
}
