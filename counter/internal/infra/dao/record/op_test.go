package record

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
	testRepo  *Repo
	ctx       = context.TODO()
	testRedis *redis.Redis
	testCache *Cache
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	testRepo = New(db, nil)
	testRedis = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})

	testCache = NewCache(testRedis)
	if err := testCache.InitFunction(ctx); err != nil {
		panic(err)
	}

	m.Run()
}

func TestRepo_Insert(t *testing.T) {
	Convey("Insert", t, func() {
		for _, m := range []*Record{
			{
				BizCode: 10000,
				Uid:     2000,
				Oid:     2000,
				Act:     1,
			},
			{
				BizCode: 10000,
				Uid:     2001,
				Oid:     2001,
				Act:     1,
			},
			{
				BizCode: 10000,
				Uid:     2002,
				Oid:     2002,
				Act:     1,
			},
		} {
			err := testRepo.Insert(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_InsertUpdate(t *testing.T) {
	Convey("Insert", t, func() {
		for _, m := range []*Record{
			{
				BizCode: 10000,
				Uid:     2000,
				Oid:     2000,
				Act:     1,
			},
			{
				BizCode: 10000,
				Uid:     2001,
				Oid:     2001,
				Act:     1,
			},
			{
				BizCode: 10000,
				Uid:     2002,
				Oid:     2002,
				Act:     1,
			},
		} {
			err := testRepo.InsertUpdate(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_Update(t *testing.T) {
	Convey("Update", t, func() {
		for _, m := range []*Record{
			{
				BizCode: 10000,
				Uid:     2000,
				Oid:     2000,
				Act:     1,
			},
			{
				BizCode: 10000,
				Uid:     2001,
				Oid:     2001,
				Act:     2,
			},
			{
				BizCode: 10000,
				Uid:     2002,
				Oid:     2002,
				Act:     1,
			},
		} {
			err := testRepo.Update(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_Find(t *testing.T) {
	Convey("Find", t, func() {
		model, err := testRepo.Find(ctx, 2000, 2000, 10000)
		So(err, ShouldBeNil)
		So(model, ShouldNotBeNil)
		t.Log(model)
		So(model.Uid, ShouldEqual, 2000)
		So(model.Oid, ShouldEqual, 2000)
	})
}

func TestRepo_PageGet(t *testing.T) {
	Convey("PageGet", t, func() {
		models, err := testRepo.PageGet(ctx, 1, ActDo, 10)
		So(err, ShouldBeNil)
		So(models, ShouldNotBeNil)
		So(len(models), ShouldEqual, 10)
		for _, model := range models {
			t.Logf("%+v\n", model)
		}
	})
}

func TestRepo_PageGetByUidOrderByMtimeWithCursor(t *testing.T) {
	Convey("PageGetByUidOrderByMtimeWithCursor", t, func() {
		models, err := testRepo.PageGetByUidOrderByMtimeWithCursor(ctx, 20001, PageGetByUidOrderByMtimeParam{
			Uid:   200,
			Count: 10,
		}, PageGetByUidOrderByMtimeCursor{
			Mtime: 1757248821,
			Id:    212,
		})
		So(err, ShouldBeNil)
		So(models, ShouldNotBeNil)
		So(len(models), ShouldEqual, 10)
		for _, model := range models {
			t.Logf("%+v\n", model)
		}
	})
}
