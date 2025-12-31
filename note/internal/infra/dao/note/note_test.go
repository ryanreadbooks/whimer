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
	noteRepo    *NoteRepo
	noteExtRepo *NoteExtRepo
	noteAssetRepo *NoteAssetRepo
	ctx         = context.TODO()
	testDb      *xsql.DB
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
	noteRepo = NewNoteRepo(db)
	noteExtRepo = NewNoteExtRepo(db)
	noteAssetRepo = NewNoteAssetRepo(db)
	m.Run()
}

func TestNote_GetByCursor(t *testing.T) {
	Convey("GetByCursor", t, func() {
		res, err := noteRepo.GetPublicByCursor(ctx, 129, 15)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}

func TestNoteExt_Upsert(t *testing.T) {
	Convey("NoteExt Upsert", t, func() {
		err := noteExtRepo.Upsert(ctx, &ExtPO{
			NoteId: 100,
			Tags:   "9223372036854775807",
		})
		So(err, ShouldBeNil)
	})
}

func TestNoteExt_Get(t *testing.T) {
	Convey("NoteExt Get", t, func() {
		err := noteExtRepo.Upsert(ctx, &ExtPO{
			NoteId: 100,
			Tags:   "1",
		})
		So(err, ShouldBeNil)
		err = noteExtRepo.Upsert(ctx, &ExtPO{
			NoteId: 200,
			Tags:   "2",
		})
		So(err, ShouldBeNil)
		err = noteExtRepo.Upsert(ctx, &ExtPO{
			NoteId: 300,
			Tags:   "3",
		})
		So(err, ShouldBeNil)

		gots, err := noteExtRepo.BatchGetById(ctx, []int64{100, 300, 200})
		So(err, ShouldBeNil)
		So(len(gots), ShouldEqual, 3)

	})
}
