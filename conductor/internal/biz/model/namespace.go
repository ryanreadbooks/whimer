package model

import (
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
)

type Namespace struct {
	Id    string    `json:"id"`
	Name  string    `json:"name"`
	Ctime time.Time `json:"ctime"`
}

func NamespaceFromPO(po *dao.NamespacePO) *Namespace {
	if po == nil {
		return nil
	}

	ctime := time.UnixMilli(po.Id.UnixMill())
	return &Namespace{
		Id:    po.Id.String(),
		Name:  po.Name,
		Ctime: ctime,
	}
}
