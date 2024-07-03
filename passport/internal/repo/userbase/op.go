package userbase

import (
	"context"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	sqlFindAll      = `select uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at from user_base where uid=?`
	sqlInsertAll    = `insert into user_base(uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at) values(?,?,?,?,?,?,?,?,?,?,?)`
	sqlDel          = `delete from user_base where uid=?`
	sqlUpdateCol    = `update user_base set %s=?,update_at=? where uid=?`
	sqlFindPassSalt = `select uid,pass,salt from user_base where uid=?`
	sqlFindBasic    = `select uid,nickname,avatar,style_sign,gender,tel,email,create_at,update_at from user_base where uid=?`
)

func (r *Repo) Find(ctx context.Context, uid uint64) (*Model, error) {
	model := new(Model)
	err := r.db.QueryRowCtx(ctx, model, sqlFindAll, uid)
	return model, xsql.ConvertError(err)
}

func (r *Repo) FindPassSalt(ctx context.Context, uid uint64) (*PassSalt, error) {
	model := new(PassSalt)
	err := r.db.QueryRowCtx(ctx, model, sqlFindPassSalt, uid)
	return model, xsql.ConvertError(err)
}

func (r *Repo) FindBasic(ctx context.Context, uid uint64) (*Basic, error) {
	model := new(Basic)
	err := r.db.QueryRowCtx(ctx, model, sqlFindBasic, uid)
	return model, xsql.ConvertError(err)
}

func (r *Repo) insert(ctx context.Context, sess sqlx.Session, user *Model) error {
	now := time.Now().Unix()
	_, err := sess.ExecCtx(ctx,
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
		now,
		now)

	return xsql.ConvertError(err)
}

func (r *Repo) Insert(ctx context.Context, user *Model) error {
	return r.insert(ctx, r.db, user)
}

func (r *Repo) InsertTx(ctx context.Context, tx sqlx.Session, user *Model) error {
	return r.insert(ctx, tx, user)
}

func (r *Repo) delete(ctx context.Context, sess sqlx.Session, uid uint64) error {
	_, err := sess.ExecCtx(ctx, sqlDel, uid)
	return xsql.ConvertError(err)
}

func (r *Repo) Delete(ctx context.Context, uid uint64) error {
	return r.delete(ctx, r.db, uid)
}

func (r *Repo) DeleteTx(ctx context.Context, tx sqlx.Session, uid uint64) error {
	return r.delete(ctx, tx, uid)
}

func (r *Repo) updateCol(ctx context.Context, sess sqlx.Session, col string, val interface{}, uid uint64) error {
	statement := fmt.Sprintf(sqlUpdateCol, col)
	_, err := sess.ExecCtx(ctx, statement, val, time.Now().Unix(), uid)
	return xsql.ConvertError(err)
}

func (r *Repo) UpdateNickname(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "nickname", value, uid)
}

func (r *Repo) UpdateNicknameTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "nickname", value, uid)
}

func (r *Repo) UpdateAvatar(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "avatar", value, uid)
}

func (r *Repo) UpdateAvatarTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "avatar", value, uid)
}

func (r *Repo) UpdateStyleSign(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "style_sign", value, uid)
}

func (r *Repo) UpdateStyleSignTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "style_sign", value, uid)
}

func (r *Repo) UpdateGender(ctx context.Context, value int8, uid uint64) error {
	return r.updateCol(ctx, r.db, "gender", value, uid)
}

func (r *Repo) UpdateGenderTx(ctx context.Context, tx sqlx.Session, value int8, uid uint64) error {
	return r.updateCol(ctx, tx, "gender", value, uid)
}

func (r *Repo) UpdateTel(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "tel", value, uid)
}

func (r *Repo) UpdateTelTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "tel", value, uid)
}

func (r *Repo) UpdateEmail(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "email", value, uid)
}

func (r *Repo) UpdateEmailTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "email", value, uid)
}

func (r *Repo) UpdatePass(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "pass", value, uid)
}

func (r *Repo) UpdatePassTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "pass", value, uid)
}

func (r *Repo) UpdateSalt(ctx context.Context, value string, uid uint64) error {
	return r.updateCol(ctx, r.db, "salt", value, uid)
}

func (r *Repo) UpdateSaltTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
	return r.updateCol(ctx, tx, "salt", value, uid)
}
