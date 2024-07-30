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
		res, err := repo.GetRootReplySortByCtime(ctx, 13, 0, 10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
		res, err = repo.GetRootReplySortByCtime(ctx, 13, 10124,10)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		for _, model := range res {
			t.Logf("%+v\n", model)
		}
	})

}
