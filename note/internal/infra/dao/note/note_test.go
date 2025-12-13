package note

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	noteDao    *NoteDao
	noteExtDao *NoteExtDao
	ctx        = context.TODO()
	testDb     *xsql.DB
)

func TestMain(m *testing.M) {
	conn := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	db := xsql.New(conn)
	testDb = db
	noteDao = NewNoteDao(db, nil)
	noteExtDao = NewNoteExtDao(db)
	m.Run()
}

func TestNote_GetByCursor(t *testing.T) {
	Convey("GetByCursor", t, func() {
		res, err := noteDao.GetPublicByCursor(ctx, 129, 15)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}

func TestNote_GetRecentPost(t *testing.T) {
	Convey("GetRecentPost", t, func() {
		res, err := noteDao.GetRecentPublicPosted(ctx, 200, 3)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}

func TestNoteExt_Upsert(t *testing.T) {
	Convey("NoteExt Upsert", t, func() {
		err := noteExtDao.Upsert(ctx, &ExtPO{
			NoteId: 100,
			Tags:   "9223372036854775807",
		})
		So(err, ShouldBeNil)
	})
}

func TestNoteExt_Get(t *testing.T) {
	Convey("NoteExt Get", t, func() {
		err := noteExtDao.Upsert(ctx, &ExtPO{
			NoteId: 100,
			Tags:   "1",
		})
		So(err, ShouldBeNil)
		err = noteExtDao.Upsert(ctx, &ExtPO{
			NoteId: 200,
			Tags:   "2",
		})
		So(err, ShouldBeNil)
		err = noteExtDao.Upsert(ctx, &ExtPO{
			NoteId: 300,
			Tags:   "3",
		})
		So(err, ShouldBeNil)

		gots, err := noteExtDao.BatchGetById(ctx, []int64{100, 300, 200})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 3)

	})
}
