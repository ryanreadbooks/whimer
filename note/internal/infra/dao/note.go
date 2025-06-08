package dao

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/global"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	sqlFind                = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE id=?"
	sqlInsertAll           = "INSERT INTO note(title,`desc`,privacy,owner,create_at,update_at) VALUES(?,?,?,?,?,?)"
	sqlUpdateAll           = "UPDATE note SET title=?,`desc`=?,privacy=?,owner=?,update_at=? WHERE id=?"
	sqlDeleteById          = "DELETE FROM note WHERE id=?"
	sqlListByOwner         = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE owner=?"
	sqlListByOwnerByCursor = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE owner=? AND id<? ORDER BY create_at DESC, id DESC LIMIT ?"
	sqlGetByCursor         = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE id>=? AND privacy=? LIMIT ?"
	sqlGetRecentPosted     = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE owner=? AND privacy=? ORDER BY create_at DESC LIMIT ?"
	sqlGetLastId           = "SELECT id FROM note WHERE privacy=? ORDER BY id DESC LIMIT 1"
	sqlGetAll              = "SELECT id,title,`desc`,privacy,owner,create_at,update_at FROM note WHERE privacy=?"
	sqlGetCount            = "SELECT COUNT(*) FROM note WHERE privacy=?"
	sqlCountByUid          = "SELECT COUNT(*) FROM note WHERE owner=?"
)

type NoteDao struct {
	db    sqlx.SqlConn
	cache *redis.Redis
}

func NewNoteDao(db sqlx.SqlConn, cache *redis.Redis) *NoteDao {
	return &NoteDao{
		db:    db,
		cache: cache,
	}
}

type Note struct {
	Id       uint64 `db:"id"`
	Title    string `db:"title"`     // 标题
	Desc     string `db:"desc"`      // 描述
	Privacy  int8   `db:"privacy"`   // 公开类型
	Owner    int64  `db:"owner"`     // 笔记作者
	CreateAt int64  `db:"create_at"` // 创建时间
	UpdateAt int64  `db:"update_at"` // 更新时间
}

func (r *NoteDao) FindOne(ctx context.Context, id uint64) (*Note, error) {
	if resp, err := r.CacheGetNote(ctx, id); err == nil && resp != nil {
		return resp, nil
	}

	resp := new(Note)
	err := r.db.QueryRowCtx(ctx, resp, sqlFind, id)
	if err == nil {
		concurrent.SafeGo(func() {
			if err2 := r.CacheSetNote(context.WithoutCancel(ctx), resp); err2 != nil {
				xlog.Msg("note dao failed to set cache when finding").Extras("noteId", resp.Id).Errorx(ctx)
			}
		})
	}
	return resp, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) ListByOwner(ctx context.Context, uid int64) ([]*Note, error) {
	res := make([]*Note, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListByOwner, uid)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteDao) ListByOwnerByCursor(ctx context.Context, uid int64, cursor uint64, limit int32) ([]*Note, error) {
	res := make([]*Note, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListByOwnerByCursor, uid, cursor, limit)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

func (r *NoteDao) insert(ctx context.Context, sess sqlx.Session, note *Note) (uint64, error) {
	now := time.Now().Unix()
	res, err := sess.ExecCtx(ctx,
		sqlInsertAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		now,
		now)

	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return uint64(newId), nil
}

func (r *NoteDao) Insert(ctx context.Context, note *Note) (uint64, error) {
	return r.insert(ctx, r.db, note)
}

func (r *NoteDao) InsertTx(ctx context.Context, tx sqlx.Session, note *Note) (uint64, error) {
	return r.insert(ctx, tx, note)
}

func (r *NoteDao) update(ctx context.Context, sess sqlx.Session, note *Note) error {
	_, err := sess.ExecCtx(ctx,
		sqlUpdateAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		time.Now().Unix(),
		note.Id,
	)

	concurrent.SafeGo(func() {
		if err2 := r.CacheDelNote(context.WithoutCancel(ctx), note.Id); err2 != nil {
			xlog.Msg("note dao failed to del note cache when updating").Extras("noteId", note.Id).Errorx(ctx)
		}
	})

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) Update(ctx context.Context, note *Note) error {
	return r.update(ctx, r.db, note)
}

func (r *NoteDao) UpdateTx(ctx context.Context, tx sqlx.Session, note *Note) error {
	return r.update(ctx, tx, note)
}

func (r *NoteDao) delete(ctx context.Context, sess sqlx.Session, id uint64) error {
	_, err := sess.ExecCtx(ctx, sqlDeleteById, id)

	concurrent.SafeGo(func() {
		if err2 := r.CacheDelNote(ctx, id); err2 != nil {
			xlog.Msg("note dao failed to del note cache when deleting").Extras("noteId", id).Errorx(ctx)
		}
	})

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) Delete(ctx context.Context, id uint64) error {
	return r.delete(ctx, r.db, id)
}

func (r *NoteDao) DeleteTx(ctx context.Context, tx sqlx.Session, id uint64) error {
	return r.delete(ctx, tx, id)
}

func (r *NoteDao) GetPublicByCursor(ctx context.Context, id uint64, count int) ([]*Note, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateByCursor(ctx context.Context, id uint64, count int) ([]*Note, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPrivate)
}

func (r *NoteDao) getByCursor(ctx context.Context, id uint64, count, privacy int) ([]*Note, error) {
	var res = make([]*Note, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetByCursor, id, privacy, count)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPublicLastId(ctx context.Context) (uint64, error) {
	return r.getLastId(ctx, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateLastId(ctx context.Context) (uint64, error) {
	return r.getLastId(ctx, global.PrivacyPrivate)
}

func (r *NoteDao) getLastId(ctx context.Context, privacy int) (uint64, error) {
	var lastId uint64
	err := r.db.QueryRowCtx(ctx, &lastId, sqlGetLastId, privacy)
	return lastId, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) getAll(ctx context.Context, privacy int) ([]*Note, error) {
	var res = make([]*Note, 0, 16)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetAll, privacy)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPublicAll(ctx context.Context) ([]*Note, error) {
	return r.getAll(ctx, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateAll(ctx context.Context) ([]*Note, error) {
	return r.getAll(ctx, global.PrivacyPrivate)
}

func (r *NoteDao) GetPublicCount(ctx context.Context) (uint64, error) {
	return r.getCount(ctx, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateCount(ctx context.Context) (uint64, error) {
	return r.getCount(ctx, global.PrivacyPrivate)
}

func (r *NoteDao) getCount(ctx context.Context, privacy int) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlGetCount, privacy)
	return cnt, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPostedCountByOwner(ctx context.Context, uid int64) (uint64, error) {
	var cnt uint64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlCountByUid, uid)
	return cnt, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetRecentPublicPosted(ctx context.Context, uid int64, count int32) ([]*Note, error) {
	var res = make([]*Note, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetRecentPosted, uid, global.PrivacyPublic, count)
	return res, xerror.Wrap(err)
}
