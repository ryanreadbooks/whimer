package note

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xcache"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	noteTableName = "note"
)

var (
	noteFields    = xsql.GetFieldSlice(&NotePO{})
	noteInsFields = xsql.GetFieldSlice(&NotePO{}, "id") // 插入时不包含 id
)

type NotePO struct {
	Id       int64           `db:"id"`
	Title    string          `db:"title"`   // 标题
	Desc     string          `db:"desc"`    // 描述
	Privacy  model.Privacy   `db:"privacy"` // 公开类型
	Owner    int64           `db:"owner"`   // 笔记作者
	Ip       []byte          `db:"ip"`
	NoteType model.NoteType  `db:"note_type"` // 笔记类型
	State    model.NoteState `db:"state"`     // 状态
	CreateAt int64           `db:"create_at"` // 创建时间
	UpdateAt int64           `db:"update_at"` // 更新时间
}

func (NotePO) TableName() string {
	return noteTableName
}

func (n *NotePO) Values() []any {
	return []any{
		n.Id,
		n.Title,
		n.Desc,
		n.Privacy,
		n.Owner,
		n.Ip,
		n.NoteType,
		n.State,
		n.CreateAt,
		n.UpdateAt,
	}
}

func (n *NotePO) InsertValues() []any {
	return []any{
		n.Title,
		n.Desc,
		n.Privacy,
		n.Owner,
		n.Ip,
		n.NoteType,
		n.State,
		n.CreateAt,
		n.UpdateAt,
	}
}

type NoteDao struct {
	db    *xsql.DB
	cache *redis.Redis

	noteCache    *xcache.Cache[*NotePO]
	integerCache *xcache.Cache[int64]
}

func NewNoteDao(db *xsql.DB, cache *redis.Redis) *NoteDao {
	return &NoteDao{
		db:    db,
		cache: cache,

		noteCache:    xcache.New[*NotePO](cache),
		integerCache: xcache.New[int64](cache),
	}
}

func (r *NoteDao) GetNoteType(ctx context.Context, id int64) (model.NoteType, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("note_type").From(noteTableName).Where(sb.Equal("id", id))
	sql, args := sb.Build()
	var noteType model.NoteType
	err := r.db.QueryRowCtx(ctx, &noteType, sql, args...)
	return noteType, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) FindOneWithoutCache(ctx context.Context, id int64) (*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("id", id))
	sql, args := sb.Build()
	resp := new(NotePO)
	err := r.db.QueryRowCtx(ctx, resp, sql, args...)
	return resp, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) FindOne(ctx context.Context, id int64) (*NotePO, error) {
	if resp, err := r.CacheGetNote(ctx, id); err == nil && resp != nil {
		return resp, nil
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()
	resp := new(NotePO)
	err := r.db.QueryRowCtx(ctx, resp, sql, args...)
	if err == nil {
		concurrent.SafeGo(func() {
			if err2 := r.CacheSetNote(context.WithoutCancel(ctx), resp); err2 != nil {
				xlog.Msg("note dao failed to set cache when finding").Extras("noteId", resp.Id).Errorx(ctx)
			}
		})
	}
	return resp, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) FindOneForUpdate(ctx context.Context, id int64) (*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("id", id))
	sb.ForUpdate()

	sql, args := sb.Build()
	resp := new(NotePO)
	err := r.db.QueryRowCtx(ctx, resp, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return resp, nil
}

// 批量获取
func (r *NoteDao) BatchGet(ctx context.Context, ids []int64) (map[int64]*NotePO, error) {
	keys := make([]string, 0, len(ids))
	keysMap := make(map[string]int64, len(ids))
	for _, id := range ids {
		key := getNoteCacheKey(id)
		keys = append(keys, key)
		keysMap[key] = id
	}

	intermediate, err := r.noteCache.MGet(ctx, keys,
		xcache.WithMGetFallbackSec[*NotePO](xtime.WeekJitterSec(time.Hour)),
		xcache.WithMGetBgSet[*NotePO](true),
		xcache.WithMGetFallback(
			func(ctx context.Context, missingKeys []string) (t map[string]*NotePO, err error) {
				if len(missingKeys) == 0 {
					return
				}

				var missings []int64
				for _, k := range missingKeys {
					missings = append(missings, keysMap[k])
				}

				sb := sqlbuilder.NewSelectBuilder()
				sb.Select(noteFields...)
				sb.From(noteTableName)
				sb.Where(sb.In("id", xslice.Any(missings)...))

				sql, args := sb.Build()
				var notes []*NotePO
				err = r.db.QueryRowsCtx(ctx, &notes, sql, args...)
				if err != nil {
					return nil, xerror.Wrap(xsql.ConvertError(err))
				}

				return xslice.MakeMap(notes, func(v *NotePO) string { return getNoteCacheKey(v.Id) }), nil
			}),
	)

	if err != nil {
		return nil, err
	}

	notes := xmap.Values(intermediate)
	resp := make(map[int64]*NotePO, len(notes))
	for _, n := range notes {
		resp[n.Id] = n
	}

	return resp, nil
}

func (r *NoteDao) ListByOwner(ctx context.Context, uid int64) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("owner", uid))

	sql, args := sb.Build()
	res := make([]*NotePO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteDao) ListByOwnerByCursor(
	ctx context.Context,
	uid int64,
	cursor int64,
	limit int32) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...)
	sb.From(noteTableName)
	sb.Where(sb.Equal("owner", uid), sb.LessThan("id", cursor))
	sb.OrderByDesc("create_at")
	sb.OrderByDesc("id")
	sb.Limit(int(limit))

	sql, args := sb.Build()
	res := make([]*NotePO, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

// ATTENTION: listing is in reverse order
func (r *NoteDao) ListPublicByOwnerByCursor(
	ctx context.Context,
	uid int64,
	cursor int64,
	limit int32,
) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(
			sb.EQ("owner", uid),
			sb.LessThan("id", cursor),
			sb.EQ("privacy", model.PrivacyPublic),
		).
		OrderByDesc("create_at").OrderByDesc("id").
		Limit(int(limit))

	sql, args := sb.Build()
	res := make([]*NotePO, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

// this is for job internal use. do not use this for biz purpose
//
// ATTENTION: listing is in reverse order
func (r *NoteDao) ListPublicByCursor(ctx context.Context, cursor int64, limit int32) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(
			sb.LessThan("id", cursor),
			sb.EQ("privacy", model.PrivacyPublic),
		).
		OrderByDesc("create_at").OrderByDesc("id").
		Limit(int(limit))

	sql, args := sb.Build()
	res := make([]*NotePO, 0, limit)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

func (r *NoteDao) PageListByOwner(ctx context.Context, uid int64, page, count int32) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(sb.EQ("owner", uid)).
		OrderByDesc("create_at").OrderByDesc("id").
		Offset(int((page - 1) * count)).
		Limit(int(count))

	sql, args := sb.Build()
	res := make([]*NotePO, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return res, nil
}

func (r *NoteDao) Insert(ctx context.Context, note *NotePO) (int64, error) {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(noteTableName)
	ib.Cols(noteInsFields...)
	ib.Values(note.InsertValues()...)

	sql, args := ib.Build()
	res, err := r.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	return int64(newId), nil
}

func (r *NoteDao) Update(ctx context.Context, note *NotePO) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(noteTableName)
	ub.Set(
		ub.EQ("title", note.Title),
		ub.EQ("`desc`", note.Desc),
		ub.EQ("privacy", note.Privacy),
		ub.EQ("owner", note.Owner),
		ub.EQ("ip", note.Ip),
		ub.EQ("note_type", note.NoteType),
		ub.EQ("state", note.State),
		ub.EQ("update_at", time.Now().Unix()),
	)
	ub.Where(ub.EQ("id", note.Id))

	sql, args := ub.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)

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

func (r *NoteDao) delete(ctx context.Context, noteId int64) error {
	// 先获取 owner
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("owner").From(noteTableName).Where(sb.EQ("id", noteId))
	sql, args := sb.Build()

	var ownerId int64
	err := r.db.QueryRowCtx(ctx, &ownerId, sql, args...)
	if err != nil {
		return xerror.Wrap(xsql.ConvertError(err))
	}

	// 删除
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(noteTableName).Where(db.EQ("id", noteId))
	sql, args = db.Build()
	_, err = r.db.ExecCtx(ctx, sql, args...)

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

func (r *NoteDao) GetPublicByCursor(ctx context.Context, id int64, count int) ([]*NotePO, error) {
	return r.getByCursor(ctx, id, count, int8(model.PrivacyPublic))
}

func (r *NoteDao) GetPrivateByCursor(ctx context.Context, id int64, count int) ([]*NotePO, error) {
	return r.getByCursor(ctx, id, count, int8(model.PrivacyPrivate))
}

func (r *NoteDao) getByCursor(ctx context.Context, id int64, count int, privacy int8) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(sb.GreaterEqualThan("id", id), sb.EQ("privacy", privacy)).
		Limit(count)

	sql, args := sb.Build()
	var res = make([]*NotePO, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPublicLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, int8(model.PrivacyPublic))
}

func (r *NoteDao) GetPrivateLastId(ctx context.Context) (int64, error) {
	return r.getLastId(ctx, int8(model.PrivacyPrivate))
}

func (r *NoteDao) getLastId(ctx context.Context, privacy int8) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id").
		From(noteTableName).
		Where(sb.EQ("privacy", privacy)).
		OrderByDesc("id").
		Limit(1)

	sql, args := sb.Build()
	var lastId int64
	err := r.db.QueryRowCtx(ctx, &lastId, sql, args...)
	return lastId, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) getAllByPrivacy(ctx context.Context, privacy model.Privacy) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(sb.EQ("privacy", privacy))

	sql, args := sb.Build()
	var res = make([]*NotePO, 0, 16)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPublicAll(ctx context.Context) ([]*NotePO, error) {
	return r.getAllByPrivacy(ctx, model.PrivacyPublic)
}

func (r *NoteDao) GetPrivateAll(ctx context.Context) ([]*NotePO, error) {
	return r.getAllByPrivacy(ctx, model.PrivacyPrivate)
}

func (r *NoteDao) GetPublicCount(ctx context.Context) (int64, error) {
	return r.getCountByPrivacy(ctx, model.PrivacyPublic)
}

func (r *NoteDao) GetPrivateCount(ctx context.Context) (int64, error) {
	return r.getCountByPrivacy(ctx, model.PrivacyPrivate)
}

func (r *NoteDao) getCountByPrivacy(ctx context.Context, privacy model.Privacy) (int64, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("COUNT(*)").
		From(noteTableName).
		Where(sb.EQ("privacy", privacy))

	sql, args := sb.Build()
	var cnt int64
	err := r.db.QueryRowCtx(ctx, &cnt, sql, args...)
	return cnt, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) GetPostedCountByOwner(ctx context.Context, uid int64) (int64, error) {
	cnt, err := r.integerCache.Get(ctx,
		getNoteCountByOwnerCacheKey(uid),
		xcache.WithGetFallback(
			func(ctx context.Context) (t int64, sec int, err error) {
				sb := sqlbuilder.NewSelectBuilder()
				sb.Select("COUNT(*)").From(noteTableName).Where(sb.EQ("owner", uid))

				sql, args := sb.Build()
				var cnt int64
				err = r.db.QueryRowCtx(ctx, &cnt, sql, args...)
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
				sb := sqlbuilder.NewSelectBuilder()
				sb.Select("COUNT(*)").
					From(noteTableName).
					Where(sb.EQ("owner", uid), sb.EQ("privacy", model.PrivacyPublic))

				sql, args := sb.Build()
				var cnt int64
				err = r.db.QueryRowCtx(ctx, &cnt, sql, args...)
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

func (r *NoteDao) GetRecentPublicPosted(ctx context.Context, uid int64, count int32) ([]*NotePO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(noteFields...).
		From(noteTableName).
		Where(sb.EQ("owner", uid), sb.EQ("privacy", model.PrivacyPublic)).
		OrderByDesc("create_at").
		Limit(int(count))

	sql, args := sb.Build()
	var res = make([]*NotePO, 0, count)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	return res, xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteDao) UpdateState(ctx context.Context, noteId int64, state model.NoteState) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(noteTableName)
	ub.Set(ub.EQ("state", state))
	ub.Where(ub.EQ("id", noteId))

	sql, args := ub.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)

	// 删除缓存
	concurrent.SafeGo(func() {
		ctx := context.WithoutCancel(ctx)
		if err := r.CacheDelNote(ctx, noteId); err != nil {
			xlog.Msg("note dao failed to del note cache when updating state").
				Err(err).Extras("noteId", noteId).Errorx(ctx)
		}
	})

	return xerror.Wrap(xsql.ConvertError(err))
}
