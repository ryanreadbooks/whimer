package p2p

import (
	"context"
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	chatDao    *ChatDao
	messageDao *MsgDao
	inboxDao   *InboxDao
	ctx        = context.TODO()
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	d := xsql.New(db)
	chatDao = NewChatDao(d)
	messageDao = NewMsgDao(d)
	inboxDao = NewInboxDao(d)
	m.Run()
}
