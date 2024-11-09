package dao

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	ctx = context.TODO()
	dao *UserDao
)

func TestMain(m *testing.M) {
	dao = NewUserDao(xsql.New(sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))), nil)

	m.Run()
}

func TestInsert(t *testing.T) {
	defer dao.Delete(ctx, 1002)

	Convey("test userbase insert", t, func() {
		err := dao.Insert(ctx, &User{
			UserBase: UserBase{
				Uid:       1002,
				Nickname:  "tester",
				Avatar:    "abc",
				StyleSign: "this is test",
				Gender:    1,
				Tel:       "121111",
				Email:     "1212@qq.com",
			},
			UserSecret: UserSecret{
				Pass: "pass",
				Salt: "salt",
			},
		})

		So(err, ShouldBeNil)

		res, err := dao.FindByUid(ctx, 1002)
		So(err, ShouldBeNil)
		t.Logf("%+v\n", res)
	})
}

func TestUpdate(t *testing.T) {
	defer dao.Delete(ctx, 1002)

	Convey("test userbase update", t, func() {
		err := dao.Insert(ctx, &User{
			UserBase: UserBase{
				Uid:       1002,
				Nickname:  "tester",
				Avatar:    "abc",
				StyleSign: "this is test",
				Gender:    1,
				Tel:       "121111",
				Email:     "1212@qq.com",
			},
			UserSecret: UserSecret{
				Pass: "pass",
				Salt: "salt",
			},
		})

		So(err, ShouldBeNil)

		err = dao.UpdateNickname(ctx, "new-tester", 1002)
		So(err, ShouldBeNil)

		res, err := dao.FindByUid(ctx, 1002)
		So(err, ShouldBeNil)
		So(res.Nickname, ShouldEqual, "new-tester")
	})
}

func TestFindPassSalt(t *testing.T) {
	defer dao.Delete(ctx, 1002)

	Convey("test userbase FindPassSalt", t, func() {
		err := dao.Insert(ctx, &User{
			UserBase: UserBase{
				Uid:       1002,
				Nickname:  "tester",
				Avatar:    "abc",
				StyleSign: "this is test",
				Gender:    1,
				Tel:       "121111",
				Email:     "1212@qq.com",
			},
			UserSecret: UserSecret{
				Pass: "pass",
				Salt: "salt",
			},
		})
		So(err, ShouldBeNil)

		model, err := dao.FindPassAndSaltByUid(ctx, 1002)
		So(err, ShouldBeNil)
		So(model.Pass, ShouldEqual, "pass")
		So(model.Salt, ShouldEqual, "salt")
	})
}

func TestFindBasic(t *testing.T) {
	defer dao.Delete(ctx, 1002)

	Convey("test userbase FindBasic", t, func() {
		err := dao.Insert(ctx, &User{
			UserBase: UserBase{
				Uid:       1002,
				Nickname:  "tester",
				Avatar:    "abc",
				StyleSign: "this is test",
				Gender:    1,
				Tel:       "121111",
				Email:     "1212@qq.com",
			},
			UserSecret: UserSecret{
				Pass: "pass",
				Salt: "salt",
			},
		})
		So(err, ShouldBeNil)

		model, err := dao.FindUserBaseByUid(ctx, 1002)
		So(err, ShouldBeNil)
		So(model.Nickname, ShouldEqual, "tester")
		So(model.Avatar, ShouldEqual, "abc")
		So(model.StyleSign, ShouldEqual, "this is test")
		So(model.Gender, ShouldEqual, 1)
		So(model.Tel, ShouldEqual, "121111")
		So(model.Email, ShouldEqual, "1212@qq.com")
	})
}

func TestFindBasicIn(t *testing.T) {
	Convey("test userbase FindBasicByUids", t, func() {
		users, err := dao.FindUserBaseByUids(ctx, []uint64{1, 10002, 20001})
		So(err, ShouldBeNil)
		So(users, ShouldNotBeEmpty)
	})
}
