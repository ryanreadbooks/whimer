package dao

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/system"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Dao struct {
	db *xsql.DB

	SystemChatDao *system.ChatDao
	SystemMsgDao  *system.SystemMsgDao

	ChatDao          *chat.ChatDao
	MsgDao           *chat.MsgDao
	ChatMsgDao       *chat.ChatMsgDao
	MsgExtDao        *chat.MsgExtDao
	ChatMemberP2PDao *chat.ChatMemberP2PDao
	ChatInboxDao     *chat.ChatInboxDao
}

func MustNew(c *config.Config) *Dao {
	conn := sqlx.NewMysql(xsql.GetDsn(
		c.MySql.User,
		c.MySql.Pass,
		c.MySql.Addr,
		c.MySql.DbName,
	))

	// 启动时必须确保数据库有效
	rdb, err := conn.RawDB()
	if err != nil {
		panic(err)
	}
	if err = rdb.Ping(); err != nil {
		panic(err)
	}

	db := xsql.New(conn)
	return &Dao{
		db: db,

		SystemChatDao: system.NewChatDao(db),
		SystemMsgDao:  system.NewSystemMsgDao(db),

		ChatDao:          chat.NewChatDao(db),
		MsgDao:           chat.NewMsgDao(db),
		ChatMsgDao:       chat.NewChatMsgDao(db),
		MsgExtDao:        chat.NewMsgExtDao(db),
		ChatMemberP2PDao: chat.NewChatMemberP2PDao(db),
		ChatInboxDao:     chat.NewChatInboxDao(db),
	}
}

func (d *Dao) DB() *xsql.DB {
	return d.db
}
