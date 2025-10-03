package dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	xcachev2 "github.com/ryanreadbooks/whimer/misc/xcache/v2"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/ryanreadbooks/whimer/passport/internal/model/consts"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type UserDao struct {
	db    *xsql.DB
	cache *redis.Redis
}

func NewUserDao(db *xsql.DB, c *redis.Redis) *UserDao {
	return &UserDao{
		db:    db,
		cache: c,
	}
}

type User struct {
	UserBase
	UserSecret
}

type UserBase struct {
	Uid       int64             `db:"uid" json:"uid,omitempty"`
	Nickname  string            `db:"nickname" json:"nickname,omitempty"`
	Avatar    string            `db:"avatar" json:"avatar,omitempty"`
	StyleSign string            `db:"style_sign" json:"style_sign,omitempty"`
	Gender    int8              `db:"gender" json:"gender,omitempty"`
	Tel       string            `db:"tel" json:"tel,omitempty"`
	Email     string            `db:"email" json:"email,omitempty"`
	Status    consts.UserStatus `db:"status" json:"status"`
	Timing
}

type UserSecret struct {
	Pass string `db:"pass" json:"-"`
	Salt string `db:"salt" json:"-"`
}

// 仅有部分查询结果的定义
type PassAndSalt struct {
	Uid  int64  `db:"uid"`
	Pass string `db:"pass"`
	Salt string `db:"salt"`
}

type Timing struct {
	CreateAt int64 `db:"create_at" json:"create_at,omitempty"`
	UpdateAt int64 `db:"update_at" json:"update_at,omitempty"`
}

// all sqls here
const (
	allFields   = "uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,status,create_at,update_at"
	basicFields = "uid,nickname,avatar,style_sign,gender,tel,email,status,create_at,update_at"

	sqlFindAll         = "SELECT " + allFields + " FROM user WHERE %s=?"
	sqlInsertAll       = "INSERT INTO user(" + allFields + ") VALUES(?,?,?,?,?,?,?,?,?,?,?,?)"
	sqlDel             = "DELETE FROM user WHERE uid=?"
	sqlUpdateCol       = "UPDATE user set %s=?,update_at=? WHERE uid=?"
	sqlFindPassSalt    = "SELECT uid,pass,salt FROM user WHERE uid=?"
	sqlFindBasic       = "SELECT " + basicFields + " FROM user WHERE %s=?"
	sqlFindBasicIn     = "SELECT " + basicFields + " FROM user WHERE uid IN (%s)"
	sqlUpdateBasicCore = "UPDATE user SET nickname=?,style_sign=?,gender=?,update_at=? WHERE uid=?"
)

func (d *UserDao) find(ctx context.Context, cond string, val any) (*User, error) {
	model := new(User)
	err := d.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindAll, cond), val)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query %s=%v failed", cond, val)
}

func (d *UserDao) FindByUid(ctx context.Context, uid int64) (*User, error) {
	return d.find(ctx, "uid", uid)
}

func (d *UserDao) FindByTel(ctx context.Context, tel string) (*User, error) {
	return d.find(ctx, "tel", tel)
}

func (d *UserDao) FindPassAndSaltByUid(ctx context.Context, uid int64) (*PassAndSalt, error) {
	model := new(PassAndSalt)
	err := d.db.QueryRowCtx(ctx, model, sqlFindPassSalt, uid)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query pass and salt failed")
}

func (d *UserDao) findUserBaseBy(ctx context.Context, cond string, val any) (*UserBase, error) {
	model := new(UserBase)
	err := d.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindBasic, cond), val)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query user base %s=%v failed", cond, val)
}

func (d *UserDao) FindUserBaseByUid(ctx context.Context, uid int64) (*UserBase, error) {
	userBase, err := xcachev2.New[*UserBase](d.cache).GetOrFetch(ctx,
		getCacheUserBaseUidKey(uid),
		func(ctx context.Context) (*UserBase, time.Duration, error) {
			dbUser, err := d.findUserBaseBy(ctx, "uid", uid)
			if err != nil {
				if errors.Is(err, xsql.ErrNoRecord) {
					// 返回假数据
					return &UserBase{Status: consts.UserStatusUnknown}, 0, nil
				}
				return nil, 0, err
			}

			return dbUser, 0, nil
		},
		xcachev2.WithTTL(xtime.WeekJitter(time.Minute*30)),
	)

	if userBase.Status.Unknown() {
		return nil, xsql.ErrNoRecord
	}

	return userBase, err
}

func (d *UserDao) FindUserBaseByTel(ctx context.Context, tel string) (*UserBase, error) {
	return d.findUserBaseBy(ctx, "tel", tel)
}

// 返回的切片不按照入参uids的顺序
func (d *UserDao) FindUserBaseByUids(ctx context.Context, uids []int64) ([]*UserBase, error) {
	if len(uids) == 0 {
		return make([]*UserBase, 0), nil
	}

	cacheKeys, keysMapping := xcachev2.KeysAndMap(uids, getCacheUserBaseUidKey)

	result, err := xcachev2.New[*UserBase](d.cache).MGetOrFetch(ctx,
		cacheKeys,
		func(ctx context.Context, keys []string) (map[string]*UserBase, error) {
			dbUids := xcachev2.RangeKeys(keys, keysMapping)
			dbUids = xslice.Uniq(dbUids)

			dbResult := make([]*UserBase, 0)
			sql := fmt.Sprintf(sqlFindBasicIn, xslice.JoinInts(dbUids))
			err := d.db.QueryRowsCtx(ctx, &dbResult, sql)
			if err != nil {
				return nil, xerror.Wrapf(xsql.ConvertError(err), "user dao query user base by uids failed")
			}

			ret := xslice.MakeMap(dbResult, func(v *UserBase) string {
				return getCacheUserBaseUidKey(v.Uid)
			})

			// 没找到的填入假数据
			for _, k := range keys {
				if _, ok := ret[k]; !ok {
					ret[k] = &UserBase{Status: consts.UserStatusUnknown}
				}
			}

			return ret, nil
		},
		xcachev2.WithTTL(xtime.WeekJitter(time.Minute*60)),
	)

	if err != nil {
		return nil, err
	}

	// filter unknown status
	return xmap.ValuesFilter(result, func(v *UserBase) bool {
		return v.Status.Unknown()
	}), nil
}

func (d *UserDao) Insert(ctx context.Context, user *User) error {
	_, err := d.db.ExecCtx(ctx,
		sqlInsertAll,
		user.Uid,
		user.Nickname,
		user.Avatar,
		user.StyleSign,
		user.Gender,
		user.Tel,
		user.Email,
		user.Pass,
		user.Salt,
		user.Status,
		user.CreateAt,
		user.UpdateAt)

	return xerror.Wrapf(xsql.ConvertError(err), "user dao insert user failed")
}

func (d *UserDao) Delete(ctx context.Context, uid int64) error {
	_, err := d.db.ExecCtx(ctx, sqlDel, uid)
	d.CacheDelUserBaseByUid(ctx, uid)
	return xerror.Wrapf(xsql.ConvertError(err), "user dao delete user %d failed", uid)
}

func (d *UserDao) updateCol(ctx context.Context, col string, val interface{}, uid int64) error {
	statement := fmt.Sprintf(sqlUpdateCol, col)
	_, err := d.db.ExecCtx(ctx, statement, val, time.Now().Unix(), uid)

	d.CacheDelUserBaseByUid(ctx, uid)

	return xerror.Wrapf(xsql.ConvertError(err), "user dao update col(%s) failed", col)
}

func (d *UserDao) UpdateNickname(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "nickname", value, uid)
}

func (d *UserDao) UpdateAvatar(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "avatar", value, uid)
}

func (d *UserDao) UpdateStyleSign(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "style_sign", value, uid)
}

func (d *UserDao) UpdateGender(ctx context.Context, value int8, uid int64) error {
	return d.updateCol(ctx, "gender", value, uid)
}

func (d *UserDao) UpdateTel(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "tel", value, uid)
}
func (d *UserDao) UpdateEmail(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "email", value, uid)
}

func (d *UserDao) UpdatePass(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "pass", value, uid)
}

func (d *UserDao) UpdateSalt(ctx context.Context, value string, uid int64) error {
	return d.updateCol(ctx, "salt", value, uid)
}

func (d *UserDao) UpdateStatus(ctx context.Context, status int8, uid int64) error {
	return d.updateCol(ctx, "status", status, uid)
}

func (d *UserDao) UpdateUserBase(ctx context.Context, base *UserBase) error {
	_, err := d.db.ExecCtx(ctx,
		sqlUpdateBasicCore,
		base.Nickname,
		base.StyleSign,
		base.Gender,
		time.Now().Unix(),
		base.Uid,
	)
	d.CacheDelUserBaseByUid(ctx, base.Uid)

	return xerror.Wrapf(xsql.ConvertError(err), "user dao update user base failed")
}
