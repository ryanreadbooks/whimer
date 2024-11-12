package dao

import (
	"context"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
)

var (
	relationDao *RelationDao
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	relationDao = NewRelationDao(xsql.NewFromEnv(), nil)
	m.Run()
}
