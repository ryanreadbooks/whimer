package system

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	systemChatDao *SystemChatDao
	systemMsgDao  *SystemMsgDao
	ctx           = context.TODO()
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	d := xsql.New(db)
	systemChatDao = NewSystemChatDao(d)
	systemMsgDao = NewSystemMsgDao(d)
	m.Run()

	// deleteForTest()
}
