package summary

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

	repo = New(db, nil)
	m.Run()
	// repo.db.Exec("DELETE FROM counter_record")
}

func TestSummaryRepo_Insert(t *testing.T) {
	Convey("Insert", t, func() {
		for _, m := range []*Model{
			{
				BizCode: 1000,
				Oid:     1000,
				Cnt:     1219,
			},
			{
				BizCode: 1000,
				Oid:     12003,
				Cnt:     1218,
			},
			{
				BizCode: 1000,
				Oid:     112,
				Cnt:     12,
			},
		} {
			err := repo.Insert(ctx, m)
			So(err, ShouldBeNil)
		}
	})
}

func TestSummaryRepo_Get(t *testing.T) {
	Convey("Get", t, func() {
		for _, m := range []struct {
			Biz int
			Oid int64
		}{
			{Biz: 1000, Oid: 1000},
			{Biz: 1000, Oid: 12003},
			{Biz: 1000, Oid: 112},
		} {
			cnt, err := repo.Get(ctx, m.Biz, m.Oid)
			So(err, ShouldBeNil)
			t.Log(cnt)
		}
	})
}

func TestSummaryRepo_Gets(t *testing.T) {
	Convey("Gets", t, func() {
		result, err := repo.Gets(ctx, []PrimaryKey{
			{BizCode: 1000, Oid: 112},
			{BizCode: 1000, Oid: 1000},
			{BizCode: 1000, Oid: 12003},
		})
		So(err, ShouldBeNil)
		t.Log(result)
	})
}

func TestSummaryRepo_BatchInsert(t *testing.T) {
	Convey("BatchInsert", t, func() {
		err := repo.BatchInsert(ctx, []*Model{
			{
				BizCode: 1000,
				Oid:     2349,
				Cnt:     123,
			},
			{
				BizCode: 1000,
				Oid:     2349,
				Cnt:     124,
			},
			{
				BizCode: 1000,
				Oid:     567,
				Cnt:     4561,
			},
			{
				BizCode: 1000,
				Oid:     4561,
				Cnt:     441,
			},
			{
				BizCode: 1000,
				Oid:     3245,
				Cnt:     157,
			},
		})
		So(err, ShouldBeNil)
	})
}

func TestSummaryRepo_Incr(t *testing.T) {
	Convey("Incr", t, func() {
		err := repo.InsertOrIncr(ctx, 1000, 112)
		So(err, ShouldBeNil)
	})
}

func TestSummaryRepo_Decr(t *testing.T) {
	Convey("Incr", t, func() {
		err := repo.InsertOrDecr(ctx, 1000, 112)
		So(err, ShouldBeNil)
	})
}
