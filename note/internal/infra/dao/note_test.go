package dao

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	dao *NoteDao
	ctx = context.TODO()
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	dao = NewNoteDao(db, nil)
	m.Run()
}

func TestNote_GetByCursor(t *testing.T) {
	Convey("GetByCursor", t, func() {
		res, err := dao.GetPublicByCursor(ctx, 129, 15)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}

func TestNote_GetRecentPost(t *testing.T) {
	Convey("GetRecentPost", t, func() {
		res, err := dao.GetRecentPublicPosted(ctx, 200, 3)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}
