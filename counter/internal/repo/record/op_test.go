package record

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
	// repo.db.Exec("DELETE FROM counter_record")
}

func TestRepo_Insert(t *testing.T) {
	Convey("Insert", t, func() {
		for _, m := range []*Model{
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
			err := repo.Insert(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_InsertUpdate(t *testing.T) {
	Convey("Insert", t, func() {
		for _, m := range []*Model{
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
			err := repo.InsertUpdate(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_Update(t *testing.T) {
	Convey("Update", t, func() {
		for _, m := range []*Model{
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
			err := repo.Update(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestRepo_Find(t *testing.T) {
	Convey("Find", t, func() {
		model, err := repo.Find(ctx, 2000, 2000, 10000)
		So(err, ShouldBeNil)
		So(model, ShouldNotBeNil)
		t.Log(model)
		So(model.Uid, ShouldEqual, 2000)
		So(model.Oid, ShouldEqual, 2000)
	})
}
