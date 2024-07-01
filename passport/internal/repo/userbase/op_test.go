package userbase_test

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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

	if err != nil {
		t.Fatal(err)
	}

	res, err := repo.Find(ctx, 1002)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%+v\n", res)
}

func TestUpdate(t *testing.T) {
	defer repo.Delete(ctx, 1002)
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

	if err != nil {
		t.Fatal(err)
	}

	err = repo.UpdateNickname(ctx, "new-tester", 1002)
	if err != nil {
		t.Fatal(err)
	}

	res, err := repo.Find(ctx, 1002)
	if err != nil {
		t.Fatal(err)
	}

	if res.Nickname != "new-tester" {
		t.Fatal("update fail")
	}

}
