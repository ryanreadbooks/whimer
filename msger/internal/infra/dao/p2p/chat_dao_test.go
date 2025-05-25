package p2p

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	chatDao *ChatDao
	ctx     = context.TODO()
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	chatDao = NewChatDao(xsql.New(db))
	m.Run()
}

func TestDao_Create(t *testing.T) {
	Convey("TestDao_Create", t, func() {
		id, err := chatDao.Create(ctx, &Chat{
			ChatId:      100,
			UserId:      10,
			PeerId:      20,
			UnReadCount: 100,
		})
		So(err, ShouldBeNil)
		So(id, ShouldNotBeZeroValue)
	})
}

func TestDao_InitChat(t *testing.T) {
	Convey("TestDao_InitChat", t, func() {
		err := chatDao.InitChat(ctx, 900, 1000, 300)
		So(err, ShouldBeNil)
	})
}
