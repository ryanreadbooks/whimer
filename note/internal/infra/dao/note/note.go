package note

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/ryanreadbooks/whimer/note/internal/global"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

// all sqls here
const (
	noteFields    = "id,title,`desc`,privacy,owner,ip,note_type,create_at,update_at"
	noteInsFields = "title,`desc`,privacy,owner,ip,note_type,create_at,update_at"

	sqlFind                   = "SELECT " + noteFields + " FROM note WHERE id=?"
	sqlInsertAll              = "INSERT INTO note(" + noteInsFields + ") VALUES(?,?,?,?,?,?,?)"
	sqlUpdateAll              = "UPDATE note SET title=?,`desc`=?,privacy=?,owner=?,ip=?,update_at=? WHERE id=?"
	sqlDeleteById             = "DELETE FROM note WHERE id=?"
	sqlListByOwner            = "SELECT " + noteFields + " FROM note WHERE owner=?"
	sqlListByOwnerByCursor    = "SELECT " + noteFields + " FROM note WHERE owner=? AND id<? ORDER BY create_at DESC, id DESC LIMIT ?"
	sqlGetByCursor            = "SELECT " + noteFields + " FROM note WHERE id>=? AND privacy=? LIMIT ?"
	sqlGetRecentPosted        = "SELECT " + noteFields + " FROM note WHERE owner=? AND privacy=? ORDER BY create_at DESC LIMIT ?"
	sqlGetLastId              = "SELECT id FROM note WHERE privacy=? ORDER BY id DESC LIMIT 1"
	sqlGetAll                 = "SELECT " + noteFields + " FROM note WHERE privacy=?"
	sqlGetCount               = "SELECT COUNT(*) FROM note WHERE privacy=?"
	sqlCountByUid             = "SELECT COUNT(*) FROM note WHERE owner=?"
	sqlBatchGet               = "SELECT " + noteFields + " FROM note WHERE id IN (%s)"
	sqlCountPublicByUid       = "SELECT COUNT(*) FROM note WHERE owner=? AND privacy=?"
	sqlFindOwnerById          = "SELECT owner FROM note WHERE id=?"
	sqlPageListPublicByCursor = "SELECT " + noteFields + " FROM note " +
		"WHERE id<? AND privacy=? ORDER BY create_at DESC, id DESC LIMIT ?"
	sqlPageListByOwner           = "SELECT " + noteFields + " FROM note WHERE owner=? ORDER BY create_at DESC, id DESC LIMIT ?,?"
	sqlListPublicByOwnerByCursor = "SELECT " + noteFields + " FROM note " +
		"WHERE owner=? AND id<? AND privacy=? " +
		"ORDER BY create_at DESC, id DESC LIMIT ?"
)

type NoteDao struct {
	db    *xsql.DB
	cache *redis.Redis

	noteCache    *xcache.Cache[*Note]
	integerCache *xcache.Cache[int64]
}

func NewNoteDao(db *xsql.DB, cache *redis.Redis) *NoteDao {
	return &NoteDao{
		db:    db,
		cache: cache,

		noteCache:    xcache.New[*Note](cache),
		integerCache: xcache.New[int64](cache),
	}
}

type Note struct {
	Id       int64  `db:"id"`
	Title    string `db:"title"`   // 标题
	Desc     string `db:"desc"`    // 描述
	Privacy  int8   `db:"privacy"` // 公开类型
	Owner    int64  `db:"owner"`   // 笔记作者
	Ip       []byte `db:"ip"`
	NoteType int8   `db:"note_type"` // 笔记类型
	CreateAt int64  `db:"create_at"` // 创建时间
	UpdateAt int64  `db:"update_at"` // 更新时间
}

func (r *NoteDao) FindOne(ctx context.Context, id int64) (*Note, error) {
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

// 批量获取
func (r *NoteDao) BatchGet(ctx context.Context, ids []int64) (map[int64]*Note, error) {
	keys := make([]string, 0, len(ids))
	keysMap := make(map[string]int64, len(ids))
	for _, id := range ids {
		key := getNoteCacheKey(id)
		keys = append(keys, key)
		keysMap[key] = id
	}

	intermediate, err := r.noteCache.MGet(ctx, keys,
		xcache.WithMGetFallbackSec[*Note](xtime.WeekJitterSec(time.Hour)),
		xcache.WithMGetBgSet[*Note](true),
		xcache.WithMGetFallback(func(ctx context.Context, missingKeys []string) (t map[string]*Note, err error) {
			if len(missingKeys) == 0 {
				return
			}

			var (
				notes    []*Note
				missings []int64
			)

			for _, k := range missingKeys {
				missings = append(missings, keysMap[k])
			}

			err = r.db.QueryRowsCtx(ctx, &notes, fmt.Sprintf(sqlBatchGet, xslice.JoinInts(ids)))
			if err != nil {
				return nil, xerror.Wrap(xsql.ConvertError(err))
			}

			return xslice.MakeMap(notes, func(v *Note) string { return getNoteCacheKey(v.Id) }), nil
		}),
	)

	if err != nil {
		return nil, err
	}

	notes := xmap.Values(intermediate)
	resp := make(map[int64]*Note, len(notes))
	for _, n := range notes {
		resp[n.Id] = n
	}

	return resp, nil
}

func (r *NoteDao) ListByOwner(ctx context.Context, uid int64) ([]*Note, error) {
	res := make([]*Note, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListByOwner, uid)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteDao) ListByOwnerByCursor(ctx context.Context, uid int64, cursor int64, limit int32) ([]*Note, error) {
	res := make([]*Note, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListByOwnerByCursor, uid, cursor, limit)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

// ATTENTION: listing is in reverse order
func (r *NoteDao) ListPublicByOwnerByCursor(ctx context.Context, uid int64, cursor int64, limit int32) ([]*Note, error) {
	res := make([]*Note, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sqlListPublicByOwnerByCursor, uid, cursor, global.PrivacyPublic, limit)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

// this is for job internal use. do not use this for biz purpose
//
// ATTENTION: listing is in reverse order
func (r *NoteDao) ListPublicByCursor(ctx context.Context, cursor int64, limit int32) ([]*Note, error) {
	res := make([]*Note, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sqlPageListPublicByCursor, cursor, global.PrivacyPublic, limit)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

func (r *NoteDao) PageListByOwner(ctx context.Context, uid int64, page, count int32) ([]*Note, error) {
	res := make([]*Note, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlPageListByOwner, uid, (page-1)*count, count)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

func (r *NoteDao) insert(ctx context.Context, note *Note) (int64, error) {
	now := time.Now().Unix()
	res, err := r.db.ExecCtx(ctx,
		sqlInsertAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		note.Ip,
		note.NoteType,
		now,
		now)

	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return int64(newId), nil
}

func (r *NoteDao) Insert(ctx context.Context, note *Note) (int64, error) {
	return r.insert(ctx, note)
}

func (r *NoteDao) update(ctx context.Context, note *Note) error {
	_, err := r.db.ExecCtx(ctx,
		sqlUpdateAll,
		note.Title,
		note.Desc,
		note.Privacy,
		note.Owner,
		note.Ip,
		time.Now().Unix(),
		note.Id,
	)

	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := r.CacheDelNote(ctx, note.Id); err != nil {
			xlog.Msg("note dao failed to del note cache when updating").
				Extras("noteId", note.Id).Err(err).Errorx(ctx)
		}

		if err := r.DelKeys(ctx,
			getNoteCountByOwnerCacheKey(note.Owner),
			getNotePublicCountByOwnerCacheKey(note.Owner)); err != nil {
			xlog.Msg("note dao failed to del note count cache when updating").
				Extras("noteId", note.Id).Err(err).Errorx(ctx)
		}
	})

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) Update(ctx context.Context, note *Note) error {
	return r.update(ctx, note)
}

func (r *NoteDao) delete(ctx context.Context, noteId int64) error {
	var ownerId int64
	err := r.db.QueryRowCtx(ctx, &ownerId, sqlFindOwnerById, noteId)
	if err != nil {
		return xerror.Wrap(xsql.ConvertError(err))
	}
	_, err = r.db.ExecCtx(ctx, sqlDeleteById, noteId)

	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := r.CacheDelNote(ctx, noteId); err != nil {
			xlog.Msg("note dao failed to del note cache when deleting").
				Err(err).Extras("noteId", noteId).Errorx(ctx)
		}

		if err := r.DelKeys(ctx,
			getNoteCountByOwnerCacheKey(ownerId),
			getNotePublicCountByOwnerCacheKey(ownerId)); err != nil {
			xlog.Msg("note dao failed to del note count cache when deleting").
				Err(err).Extras("noteId", noteId).Errorx(ctx)
		}
	})

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) Delete(ctx context.Context, id int64) error {
	return r.delete(ctx, id)
}

func (r *NoteDao) GetPublicByCursor(ctx context.Context, id int64, count int) ([]*Note, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateByCursor(ctx context.Context, id int64, count int) ([]*Note, error) {
	return r.getByCursor(ctx, id, count, global.PrivacyPrivate)
}

func (r *NoteDao) getByCursor(ctx context.Context, id int64, count int, privacy int8) ([]*Note, error) {
	var res = make([]*Note, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetByCursor, id, privacy, count)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPublicLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, global.PrivacyPrivate)
}

func (r *NoteDao) getLastId(ctx context.Context, privacy int8) (int64, error) {
	var lastId int64
	err := r.db.QueryRowCtx(ctx, &lastId, sqlGetLastId, privacy)
	return lastId, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) getAll(ctx context.Context, privacy int8) ([]*Note, error) {
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

func (r *NoteDao) GetPublicCount(ctx context.Context) (int64, error) {
	return r.getCount(ctx, global.PrivacyPublic)
}

func (r *NoteDao) GetPrivateCount(ctx context.Context) (int64, error) {
	return r.getCount(ctx, global.PrivacyPrivate)
}

func (r *NoteDao) getCount(ctx context.Context, privacy int8) (int64, error) {
	var cnt int64
	err := r.db.QueryRowCtx(ctx, &cnt, sqlGetCount, privacy)
	return cnt, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPostedCountByOwner(ctx context.Context, uid int64) (int64, error) {
	cnt, err := r.integerCache.Get(ctx,
		getNoteCountByOwnerCacheKey(uid),
		xcache.WithGetFallback(
			func(ctx context.Context) (t int64, sec int, err error) {
				var cnt int64
				err = r.db.QueryRowCtx(ctx, &cnt, sqlCountByUid, uid)
				if err != nil {
					return 0, 0, xerror.Wrap(xsql.ConvertError(err))
				}

				return cnt, xtime.WeekJitterSec(time.Hour * 2), nil
			},
		),
	)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (r *NoteDao) GetPublicPostedCountByOwner(ctx context.Context, uid int64) (int64, error) {
	cnt, err := r.integerCache.Get(ctx,
		getNotePublicCountByOwnerCacheKey(uid),
		xcache.WithGetFallback(
			func(ctx context.Context) (t int64, sec int, err error) {
				var cnt int64
				err = r.db.QueryRowCtx(ctx, &cnt, sqlCountPublicByUid, uid, global.PrivacyPublic)
				if err != nil {
					return 0, 0, xerror.Wrap(xsql.ConvertError(err))
				}

				return cnt, xtime.WeekJitterSec(time.Hour * 2), nil
			},
		),
	)
	if err != nil {
		return 0, err
	}

	return cnt, nil
}

func (r *NoteDao) GetRecentPublicPosted(ctx context.Context, uid int64, count int32) ([]*Note, error) {
	var res = make([]*Note, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sqlGetRecentPosted, uid, global.PrivacyPublic, count)
	return res, xerror.Wrap(xsql.ConvertError(err))
}
