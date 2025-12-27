package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/uuid"
)

const (
	namespacePOTableName = "conductor_namespace"
)

var (
	namespacePOFields = xsql.GetFieldSlice(&NamespacePO{})
)

type NamespacePO struct {
	Id   uuid.UUID `db:"id"   json:"id"`   
	Name string `db:"name" json:"name"` 
}

func (NamespacePO) TableName() string {
	return namespacePOTableName
}

func (s *NamespacePO) Values() []any {
	return []any{
		s.Id,
		s.Name,
	}
}
