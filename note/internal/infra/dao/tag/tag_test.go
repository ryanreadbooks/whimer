package tag

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	dao *TagDao
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	conn := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	dao = NewTagDao(xsql.New(conn))
	m.Run()
}

func TestInsert(t *testing.T) {
	Convey("TestInsert", t, func() {

		tag := Tag{
			Name: "name",
		}
		id, err := dao.Create(ctx, &tag)
		So(err, ShouldBeNil)

		got, err := dao.FindById(ctx, id)
		So(err, ShouldBeNil)
		So(got.Name, ShouldEqual, tag.Name)
	})
}
