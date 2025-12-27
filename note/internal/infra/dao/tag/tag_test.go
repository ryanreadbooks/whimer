package tag

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	repo  *TagRepo
	cache *redis.Redis
	ctx   = context.TODO()
)

func TestMain(m *testing.M) {
	conn := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))
	cache = redis.MustNewRedis(redis.RedisConf{
		Host: "127.0.0.1:7542",
		Type: "node",
	})
	repo = NewTagRepo(xsql.New(conn))
	m.Run()
}

func TestInsert(t *testing.T) {
	Convey("TestInsert", t, func() {

		tag := Tag{
			Name: "name",
		}
		id, err := repo.Create(ctx, &tag)
		So(err, ShouldBeNil)

		got, err := repo.FindById(ctx, id)
		So(err, ShouldBeNil)
		So(got.Name, ShouldEqual, tag.Name)
	})
}

func TestRedisMget(t *testing.T) {
	res, err := cache.MgetCtx(ctx, "abc", "efg", "qw")
	t.Log(err)
	t.Log(res)
}
