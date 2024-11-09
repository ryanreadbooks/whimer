package repo

// import (
// 	"github.com/ryanreadbooks/whimer/misc/xsql"
// 	"github.com/ryanreadbooks/whimer/passport/internal/config"
// 	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

// 	"github.com/zeromicro/go-zero/core/stores/sqlx"
// )

// type Repo struct {
// 	db sqlx.SqlConn

// 	UserBaseRepo *userbase.Repo
// }

// func New(c *config.Config) *Repo {
// 	db := sqlx.NewMysql(xsql.GetDsn(
// 		c.MySql.User,
// 		c.MySql.Pass,
// 		c.MySql.Addr,
// 		c.MySql.DbName,
// 	))

// 	// 启动时必须确保数据库有效
// 	rdb, err := db.RawDB()
// 	if err != nil {
// 		panic(err)
// 	}
// 	if err = rdb.Ping(); err != nil {
// 		panic(err)
// 	}

// 	return &Repo{
// 		db:           db,
// 		UserBaseRepo: userbase.New(db),
// 	}
// }

// func (d *Repo) DB() sqlx.SqlConn {
// 	return d.db
// }
