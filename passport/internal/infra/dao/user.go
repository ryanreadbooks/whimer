package dao

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
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
	Uid       uint64 `db:"uid" json:"uid"`
	Nickname  string `db:"nickname" json:"nickname"`
	Avatar    string `db:"avatar" json:"avatar"`
	StyleSign string `db:"style_sign" json:"style_sign"`
	Gender    int8   `db:"gender" json:"gender"`
	Tel       string `db:"tel" json:"tel"`
	Email     string `db:"email" json:"email"`
	Timing
}

type UserSecret struct {
	Pass string `db:"pass" json:"-"`
	Salt string `db:"salt" json:"-"`
}

// 仅有部分查询结果的定义
type PassAndSalt struct {
	Uid  uint64 `db:"uid"`
	Pass string `db:"pass"`
	Salt string `db:"salt"`
}

type Timing struct {
	CreateAt int64 `db:"create_at" json:"create_at,omitempty"`
	UpdateAt int64 `db:"update_at" json:"update_at,omitempty"`
}

// all sqls here
const (
	sqlFindAll         = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at FROM user WHERE %s=?`
	sqlInsertAll       = `INSERT INTO user(uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
	sqlDel             = `DELETE FROM user WHERE uid=?`
	sqlUpdateCol       = `UPDATE user set %s=?,update_at=? WHERE uid=?`
	sqlFindPassSalt    = `SELECT uid,pass,salt FROM user WHERE uid=?`
	sqlFindBasic       = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,create_at,update_at FROM user WHERE %s=?`
	sqlFindBasicIn     = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,create_at,update_at FROM user WHERE uid IN (%s)`
	sqlUpdateBasicCore = `UPDATE user SET nickname=?,style_sign=?,gender=?,update_at=? WHERE uid=?`
)

func (d *UserDao) find(ctx context.Context, cond string, val interface{}) (*User, error) {
	model := new(User)
	err := d.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindAll, cond), val)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query %s=%v failed", cond, val)
}

func (d *UserDao) FindByUid(ctx context.Context, uid uint64) (*User, error) {
	return d.find(ctx, "uid", uid)
}

func (d *UserDao) FindByTel(ctx context.Context, tel string) (*User, error) {
	return d.find(ctx, "tel", tel)
}

func (d *UserDao) FindPassAndSaltByUid(ctx context.Context, uid uint64) (*PassAndSalt, error) {
	model := new(PassAndSalt)
	err := d.db.QueryRowCtx(ctx, model, sqlFindPassSalt, uid)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query pass and salt failed")
}

func (d *UserDao) findUserBaseBy(ctx context.Context, cond string, val interface{}) (*UserBase, error) {
	model := new(UserBase)
	err := d.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindBasic, cond), val)
	return model, xerror.Wrapf(xsql.ConvertError(err), "user dao query user base %s=%v failed", cond, val)
}

func (d *UserDao) FindUserBaseByUid(ctx context.Context, uid uint64) (*UserBase, error) {
	if resp, err := d.CacheGetUserBaseByUid(ctx, uid); err == nil && resp != nil {
		return resp, nil
	}

	return d.findUserBaseBy(ctx, "uid", uid)
}

func (d *UserDao) FindUserBaseByTel(ctx context.Context, tel string) (*UserBase, error) {
	// if resp, err := d.CacheGetUserBaseByTel(ctx, tel); err == nil && resp != nil {
	// 	return resp, nil
	// }

	return d.findUserBaseBy(ctx, "tel", tel)
}

// TODO make it cache
func (d *UserDao) FindUserBaseByUids(ctx context.Context, uids []uint64) ([]*UserBase, error) {
	model := make([]*UserBase, 0)
	if len(uids) == 0 {
		return model, nil
	}

	sql := fmt.Sprintf(sqlFindBasicIn, slices.JoinInts(uids))
	err := d.db.QueryRowsCtx(ctx, &model, sql)
	if err != nil {
		return nil, xerror.Wrapf(xsql.ConvertError(err), "user dao query user base by uids failed")
	}

	return model, nil
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
		user.CreateAt,
		user.UpdateAt)

	return xerror.Wrapf(xsql.ConvertError(err), "user dao insert user failed")
}

func (d *UserDao) Delete(ctx context.Context, uid uint64) error {
	_, err := d.db.ExecCtx(ctx, sqlDel, uid)
	return xerror.Wrapf(xsql.ConvertError(err), "user dao delete user %d failed", uid)
}

func (d *UserDao) updateCol(ctx context.Context, col string, val interface{}, uid uint64) error {
	statement := fmt.Sprintf(sqlUpdateCol, col)
	_, err := d.db.ExecCtx(ctx, statement, val, time.Now().Unix(), uid)

	defer func() {
	}()

	return xerror.Wrapf(xsql.ConvertError(err), "user dao update col(%s) failed", col)
}

func (d *UserDao) UpdateNickname(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "nickname", value, uid)
}

func (d *UserDao) UpdateAvatar(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "avatar", value, uid)
}

func (d *UserDao) UpdateStyleSign(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "style_sign", value, uid)
}

func (d *UserDao) UpdateGender(ctx context.Context, value int8, uid uint64) error {
	return d.updateCol(ctx, "gender", value, uid)
}

func (d *UserDao) UpdateTel(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "tel", value, uid)
}
func (d *UserDao) UpdateEmail(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "email", value, uid)
}

func (d *UserDao) UpdatePass(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "pass", value, uid)
}

func (d *UserDao) UpdateSalt(ctx context.Context, value string, uid uint64) error {
	return d.updateCol(ctx, "salt", value, uid)
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

	return xerror.Wrapf(xsql.ConvertError(err), "user dao update user base failed")
}
