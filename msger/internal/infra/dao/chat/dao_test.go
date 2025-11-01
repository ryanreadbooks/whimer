package chat

import (
	"os"
	"testing"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var (
	testChatDao          *ChatDao
	testMsgDao           *MsgDao
	testChatMemberP2PDao *ChatMemberP2PDao
	testChatInboxDao     *ChatInboxDao
)

func TestMain(m *testing.M) {
	db := sqlx.NewMysql(xsql.GetDsn(
		os.Getenv("ENV_DB_USER"),
		os.Getenv("ENV_DB_PASS"),
		os.Getenv("ENV_DB_ADDR"),
		os.Getenv("ENV_DB_NAME"),
	))

	d := xsql.New(db)
	testChatDao = NewChatDao(d)
	testMsgDao = NewMsgDao(d)
	testChatMemberP2PDao = NewChatMemberP2PDao(d)
	testChatInboxDao = NewChatInboxDao(d)
	m.Run()
}
