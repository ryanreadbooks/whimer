package dao

import (
	"context"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	testRelationDao *RelationDao
	testSettingDao  *RelationSettingDao
	ctx             = context.TODO()
)

func TestMain(m *testing.M) {
	rd := redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})
	db := xsql.NewFromEnv()
	testRelationDao = NewRelationDao(db)
	testSettingDao = NewRelationSettingDao(db, rd)
	m.Run()
}
