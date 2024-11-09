package userbase

// import (
// 	"context"
// 	"fmt"
// 	"time"

// 	"github.com/ryanreadbooks/whimer/misc/utils/slices"
// 	"github.com/ryanreadbooks/whimer/misc/xsql"
// 	"github.com/zeromicro/go-zero/core/stores/sqlx"
// )

// // all sqls here
// const (
// 	sqlFindAll         = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at FROM user_base WHERE %s=?`
// 	sqlInsertAll       = `INSERT INTO user_base(uid,nickname,avatar,style_sign,gender,tel,email,pass,salt,create_at,update_at) VALUES(?,?,?,?,?,?,?,?,?,?,?)`
// 	sqlDel             = `DELETE FROM user_base WHERE uid=?`
// 	sqlUpdateCol       = `UPDATE user_base set %s=?,update_at=? WHERE uid=?`
// 	sqlFindPassSalt    = `SELECT uid,pass,salt FROM user_base WHERE uid=?`
// 	sqlFindBasic       = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,create_at,update_at FROM user_base WHERE %s=?`
// 	sqlFindBasicIn     = `SELECT uid,nickname,avatar,style_sign,gender,tel,email,create_at,update_at FROM user_base WHERE uid IN (%s)`
// 	sqlUpdateBasicCore = `UPDATE user_base SET nickname=?,style_sign=?,gender=?,update_at=? WHERE uid=?`
// )

// func (r *Repo) find(ctx context.Context, cond string, val interface{}) (*Model, error) {
// 	model := new(Model)
// 	err := r.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindAll, cond), val)
// 	return model, xsql.ConvertError(err)
// }

// func (r *Repo) Find(ctx context.Context, uid uint64) (*Model, error) {
// 	return r.find(ctx, "uid", uid)
// }

// func (r *Repo) FindBasicByUids(ctx context.Context, uids []uint64) ([]*Basic, error) {
// 	model := make([]*Basic, 0)
// 	if len(uids) == 0 {
// 		return model, nil
// 	}

// 	sql := fmt.Sprintf(sqlFindBasicIn, slices.JoinInts(uids))
// 	err := r.db.QueryRowsCtx(ctx, &model, sql)
// 	if err != nil {
// 		return nil, xsql.ConvertError(err)
// 	}

// 	return model, nil
// }

// func (r *Repo) FindByTel(ctx context.Context, tel string) (*Model, error) {
// 	return r.find(ctx, "tel", tel)
// }

// func (r *Repo) FindPassAndSalt(ctx context.Context, uid uint64) (*PassSalt, error) {
// 	model := new(PassSalt)
// 	err := r.db.QueryRowCtx(ctx, model, sqlFindPassSalt, uid)
// 	return model, xsql.ConvertError(err)
// }

// func (r *Repo) findBasicBy(ctx context.Context, cond string, val interface{}) (*Basic, error) {
// 	model := new(Basic)
// 	err := r.db.QueryRowCtx(ctx, model, fmt.Sprintf(sqlFindBasic, cond), val)
// 	return model, xsql.ConvertError(err)
// }

// func (r *Repo) FindBasic(ctx context.Context, uid uint64) (*Basic, error) {
// 	return r.findBasicBy(ctx, "uid", uid)
// }

// func (r *Repo) FindBasicByTel(ctx context.Context, tel string) (*Basic, error) {
// 	return r.findBasicBy(ctx, "tel", tel)
// }

// func (r *Repo) insert(ctx context.Context, sess sqlx.Session, user *Model) error {
// 	_, err := sess.ExecCtx(ctx,
// 		sqlInsertAll,
// 		user.Uid,
// 		user.Nickname,
// 		user.Avatar,
// 		user.StyleSign,
// 		user.Gender,
// 		user.Tel,
// 		user.Email,
// 		user.Pass,
// 		user.Salt,
// 		user.CreateAt,
// 		user.UpdateAt)

// 	return xsql.ConvertError(err)
// }

// func (r *Repo) Insert(ctx context.Context, user *Model) error {
// 	return r.insert(ctx, r.db, user)
// }

// func (r *Repo) InsertTx(ctx context.Context, tx sqlx.Session, user *Model) error {
// 	return r.insert(ctx, tx, user)
// }

// func (r *Repo) delete(ctx context.Context, sess sqlx.Session, uid uint64) error {
// 	_, err := sess.ExecCtx(ctx, sqlDel, uid)
// 	return xsql.ConvertError(err)
// }

// func (r *Repo) Delete(ctx context.Context, uid uint64) error {
// 	return r.delete(ctx, r.db, uid)
// }

// func (r *Repo) DeleteTx(ctx context.Context, tx sqlx.Session, uid uint64) error {
// 	return r.delete(ctx, tx, uid)
// }

// func (r *Repo) updateCol(ctx context.Context, sess sqlx.Session, col string, val interface{}, uid uint64) error {
// 	statement := fmt.Sprintf(sqlUpdateCol, col)
// 	_, err := sess.ExecCtx(ctx, statement, val, time.Now().Unix(), uid)
// 	return xsql.ConvertError(err)
// }

// func (r *Repo) UpdateNickname(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "nickname", value, uid)
// }

// func (r *Repo) UpdateNicknameTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "nickname", value, uid)
// }

// func (r *Repo) UpdateAvatar(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "avatar", value, uid)
// }

// func (r *Repo) UpdateAvatarTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "avatar", value, uid)
// }

// func (r *Repo) UpdateStyleSign(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "style_sign", value, uid)
// }

// func (r *Repo) UpdateStyleSignTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "style_sign", value, uid)
// }

// func (r *Repo) UpdateGender(ctx context.Context, value int8, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "gender", value, uid)
// }

// func (r *Repo) UpdateGenderTx(ctx context.Context, tx sqlx.Session, value int8, uid uint64) error {
// 	return r.updateCol(ctx, tx, "gender", value, uid)
// }

// func (r *Repo) UpdateTel(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "tel", value, uid)
// }

// func (r *Repo) UpdateTelTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "tel", value, uid)
// }

// func (r *Repo) UpdateEmail(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "email", value, uid)
// }

// func (r *Repo) UpdateEmailTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "email", value, uid)
// }

// func (r *Repo) UpdatePass(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "pass", value, uid)
// }

// func (r *Repo) UpdatePassTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "pass", value, uid)
// }

// func (r *Repo) UpdateSalt(ctx context.Context, value string, uid uint64) error {
// 	return r.updateCol(ctx, r.db, "salt", value, uid)
// }

// func (r *Repo) UpdateSaltTx(ctx context.Context, tx sqlx.Session, value string, uid uint64) error {
// 	return r.updateCol(ctx, tx, "salt", value, uid)
// }

// func (r *Repo) UpdateBasicCore(ctx context.Context, core *Basic) error {
// 	_, err := r.db.ExecCtx(ctx,
// 		sqlUpdateBasicCore,
// 		core.Nickname,
// 		core.StyleSign,
// 		core.Gender,
// 		time.Now().Unix(),
// 		core.Uid,
// 	)

// 	return xsql.ConvertError(err)
// }
