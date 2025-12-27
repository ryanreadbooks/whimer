package data

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	tagdao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/tag"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// Data 数据层入口
// 协调数据库和缓存操作，对上层(biz)屏蔽底层数据存取细节
type Data struct {
	db    *xsql.DB
	cache *redis.Redis

	Note            *NoteData
	NoteAsset       *NoteAssetData
	NoteExt         *NoteExtData
	ProcedureRecord *ProcedureRecordData
	Tag             *TagData
}

// MustNew 创建Data实例
func MustNew(c *config.Config, cache *redis.Redis) *Data {
	sqlx.DisableStmtLog()

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

	// 初始化底层dao - note相关
	noteRepo := notedao.NewNoteRepo(db)
	noteCache := notedao.NewNoteCache(cache)
	noteAssetRepo := notedao.NewNoteAssetRepo(db)
	noteExtRepo := notedao.NewNoteExtRepo(db)
	procedureRecordRepo := notedao.NewProcedureRecordRepo(db)

	// 初始化底层dao - tag相关
	tagRepo := tagdao.NewTagRepo(db)
	tagCache := tagdao.NewTagCache(cache)

	return &Data{
		db:              db,
		cache:           cache,
		Note:            NewNoteData(noteRepo, noteCache),
		NoteAsset:       NewNoteAssetData(noteAssetRepo),
		NoteExt:         NewNoteExtData(noteExtRepo),
		ProcedureRecord: NewProcedureRecordData(procedureRecordRepo),
		Tag:             NewTagData(tagRepo, tagCache),
	}
}

func (d *Data) DB() *xsql.DB {
	return d.db
}

func (d *Data) Close() {
	if rd, _ := d.db.Conn().RawDB(); rd != nil {
		rd.Close()
	}
}
