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

func TestNote_GetByCursor(t *testing.T) {
	Convey("GetByCursor", t, func() {
		res, err := repo.GetPublicByCursor(ctx, 129, 15)
		So(err, ShouldBeNil)
		for _, r := range res {
			t.Logf("%+v\n", r)
		}
	})
}
