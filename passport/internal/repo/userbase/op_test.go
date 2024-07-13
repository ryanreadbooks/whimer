package userbase_test

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx  = context.TODO()
	repo *userbase.Repo
)

func TestMain(m *testing.M) {
	repo = userbase.New(sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	)))

	m.Run()
}

func TestInsert(t *testing.T) {
	defer repo.Delete(ctx, 1002)

	Convey("test userbase insert", t, func() {
		err := repo.Insert(ctx, &userbase.Model{
			Uid:       1002,
			Nickname:  "tester",
			Avatar:    "abc",
			StyleSign: "this is test",
			Gender:    1,
			Tel:       "121111",
			Email:     "1212@qq.com",
			Pass:      "pass",
			Salt:      "salt",
		})

		So(err, ShouldBeNil)

		res, err := repo.Find(ctx, 1002)
		So(err, ShouldBeNil)
		t.Logf("%+v\n", res)
	})
}

func TestUpdate(t *testing.T) {
	defer repo.Delete(ctx, 1002)

	Convey("test userbase update", t, func() {
		err := repo.Insert(ctx, &userbase.Model{
			Uid:       1002,
			Nickname:  "tester",
			Avatar:    "abc",
			StyleSign: "this is test",
			Gender:    1,
			Tel:       "121111",
			Email:     "1212@qq.com",
			Pass:      "pass",
			Salt:      "salt",
		})

		So(err, ShouldBeNil)

		err = repo.UpdateNickname(ctx, "new-tester", 1002)
		So(err, ShouldBeNil)

		res, err := repo.Find(ctx, 1002)
		So(err, ShouldBeNil)
		So(res.Nickname, ShouldEqual, "new-tester")
	})
}

func TestFindPassSalt(t *testing.T) {
	defer repo.Delete(ctx, 1002)

	Convey("test userbase FindPassSalt", t, func() {
		err := repo.Insert(ctx, &userbase.Model{
			Uid:       1002,
			Nickname:  "tester",
			Avatar:    "abc",
			StyleSign: "this is test",
			Gender:    1,
			Tel:       "121111",
			Email:     "1212@qq.com",
			Pass:      "pass",
			Salt:      "salt",
		})
		So(err, ShouldBeNil)

		model, err := repo.FindPassSalt(ctx, 1002)
		So(err, ShouldBeNil)
		So(model.Pass, ShouldEqual, "pass")
		So(model.Salt, ShouldEqual, "salt")
	})
}

func TestFindBasic(t *testing.T) {
	defer repo.Delete(ctx, 1002)

	Convey("test userbase FindBasic", t, func() {
		err := repo.Insert(ctx, &userbase.Model{
			Uid:       1002,
			Nickname:  "tester",
			Avatar:    "abc",
			StyleSign: "this is test",
			Gender:    1,
			Tel:       "121111",
			Email:     "1212@qq.com",
			Pass:      "pass",
			Salt:      "salt",
		})
		So(err, ShouldBeNil)

		model, err := repo.FindBasic(ctx, 1002)
		So(err, ShouldBeNil)
		So(model.Nickname, ShouldEqual, "tester")
		So(model.Avatar, ShouldEqual, "abc")
		So(model.StyleSign, ShouldEqual, "this is test")
		So(model.Gender, ShouldEqual, 1)
		So(model.Tel, ShouldEqual, "121111")
		So(model.Email, ShouldEqual, "1212@qq.com")
	})
}
