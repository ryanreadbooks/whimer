package comm

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	repo *Repo
	ctx  = context.TODO()
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	repo = New(db)
	m.Run()
}

func TestRepo_GetRootReplySortByCtime(t *testing.T) {
	Convey("GetRootReplySortByCtime", t, func() {
		res, err := repo.GetRootReplies(ctx, 13, 0, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
		res, err = repo.GetRootReplies(ctx, 13, 10124, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
	})

}

func TestRepo_CountByOid(t *testing.T) {
	Convey("CountByOid", t, func() {
		cnt, err := repo.CountByOid(ctx, 13)
		So(err, ShouldBeNil)
		So(cnt, ShouldNotBeZeroValue)
		t.Logf("cnt = %d\n", cnt)
	})
}

func TestRepo_CountGroupByOid(t *testing.T) {
	Convey("CountGroupByOid", t, func() {
		res, err := repo.CountGroupByOid(ctx)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		t.Logf("res = %v\n", res)
	})
}

func TestRepo_CountGroupByOidLimit(t *testing.T) {
	Convey("CountGroupByOidLimit", t, func() {
		res, err := repo.CountGroupByOidLimit(ctx, 1, 2)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		t.Logf("res = %v\n", res)
	})
}
