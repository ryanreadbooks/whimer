package dao

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	testCommentDao      *CommentDao
	testCommentAssetDao *CommentAssetDao
	testCommentExtDao   *CommentExtDao
	testCtx             = context.TODO()
	testCache           *redis.Redis
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	testCache = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	testCommentDao = NewCommentDao(xsql.New(db), testCache)
	testCommentAssetDao = NewCommentAssetDao(xsql.New(db), testCache)
	testCommentExtDao = NewCommentExtDao(xsql.New(db), testCache)
	m.Run()
}

func TestRepo_GetRootCommentSortByCtime(t *testing.T) {
	Convey("GetRootCommentSortByCtime", t, func() {
		res, err := testCommentDao.GetRoots(testCtx, 13, 0, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
		res, err = testCommentDao.GetRoots(testCtx, 13, 10124, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
	})

}

func TestRepo_CountByOid(t *testing.T) {
	Convey("CountByOid", t, func() {
		cnt, err := testCommentDao.CountByOid(testCtx, 13)
		So(err, ShouldBeNil)
		So(cnt, ShouldNotBeZeroValue)
		t.Logf("cnt = %d\n", cnt)
	})
}

func TestRepo_CountGroupByOid(t *testing.T) {
	Convey("CountGroupByOid", t, func() {
		res, err := testCommentDao.CountGroupByOid(testCtx)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		t.Logf("res = %v\n", res)
	})
}

func TestRepo_CountGroupByOidLimit(t *testing.T) {
	Convey("CountGroupByOidLimit", t, func() {
		res, err := testCommentDao.CountGroupByOidLimit(testCtx, 1, 2)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		t.Logf("res = %v\n", res)
	})
}

func TestRepo_BatchCountSubReplies(t *testing.T) {
	Convey("BatchCountSubReplies", t, func() {
		res, err := testCommentDao.BatchCountSubs(testCtx, []int64{50})
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		t.Logf("res = %v\n", res)
	})
}
