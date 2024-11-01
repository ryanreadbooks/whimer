package noteasset

import (
	"context"
	"fmt"
	"strings"
	"time"

	uslices "github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// all sqls here
const (
	sqlFindAll               = `select id,asset_key,asset_type,note_id,create_at from note_asset where id=?`
	sqlInsert                = `insert into note_asset(asset_key,asset_type,note_id,create_at) values (?,?,?,?)`
	sqlFindByNoteId          = `select id,asset_key,asset_type,note_id,create_at from note_asset where note_id=?`
	sqlDeleteByNoteId        = `delete from note_asset where note_id=?`
	sqlBatchInsert           = `insert into note_asset(asset_key,asset_type,note_id,create_at) values %s`
	sqlFindByNoteIds         = `select id,asset_key,asset_type,note_id,create_at from note_asset where note_id in (%s)`
	sqlExcludeDeleteByNoteId = `delete from note_asset where note_id=?`
)

func (r *Repo) FindOne(ctx context.Context, id uint64) (*Model, error) {
	model := new(Model)
	err := r.db.QueryRowCtx(ctx, model, sqlFindAll, id)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return model, nil
}
func (r *Repo) insert(ctx context.Context, sess sqlx.Session, asset *Model) error {
	now := time.Now().Unix()
	_, err := sess.ExecCtx(ctx,
		sqlInsert,
		asset.AssetKey,
		asset.AssetType,
		asset.NoteId,
		now)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *Repo) Insert(ctx context.Context, asset *Model) error {
	return r.insert(ctx, r.db, asset)
}

func (r *Repo) InsertTx(ctx context.Context, tx sqlx.Session, asset *Model) error {
	return r.insert(ctx, tx, asset)
}

func (r *Repo) findByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64) ([]*Model, error) {
	res := make([]*Model, 0)
	err := sess.QueryRowsCtx(ctx, &res, sqlFindByNoteId, noteId)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *Repo) FindByNoteId(ctx context.Context, noteId uint64) ([]*Model, error) {
	return r.findByNoteId(ctx, r.db, noteId)
}

func (r *Repo) FindByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64) ([]*Model, error) {
	return r.findByNoteId(ctx, tx, noteId)
}

func (r *Repo) deleteByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64) error {
	_, err := sess.ExecCtx(ctx, sqlDeleteByNoteId, noteId)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *Repo) DeleteByNoteId(ctx context.Context, noteId uint64) error {
	return r.deleteByNoteId(ctx, r.db, noteId)
}

func (r *Repo) DeleteByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64) error {
	return r.deleteByNoteId(ctx, tx, noteId)
}

func (r *Repo) excludeDeleteByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64, assetKeys []string) error {
	var alen = len(assetKeys)
	var args []any = make([]any, 0, alen)
	args = append(args, noteId)

	query := sqlExcludeDeleteByNoteId
	if alen != 0 {
		var tmpl string
		for i, ask := range assetKeys {
			tmpl += "?"
			args = append(args, ask)
			if i != alen-1 {
				tmpl += ","
			}
		}
		query += fmt.Sprintf(" and `asset_key` not in (%s)", tmpl)
	}
	_, err := sess.ExecCtx(ctx, query, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}
func (r *Repo) ExcludeDeleteByNoteId(ctx context.Context, noteId uint64, assetKeys []string) error {
	return r.excludeDeleteByNoteId(ctx, r.db, noteId, assetKeys)
}

func (r *Repo) ExcludeDeleteByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64, assetKeys []string) error {
	return r.excludeDeleteByNoteId(ctx, tx, noteId, assetKeys)
}

func (r *Repo) batchInsert(ctx context.Context, sess sqlx.Session, assets []*Model) error {
	if len(assets) == 0 {
		return nil
	}

	tmpl := "(?, ?, ?, ?)"
	var builder strings.Builder
	var args []any = make([]any, 0, len(assets)*4)
	for i, data := range assets {
		builder.WriteString(tmpl)
		args = append(args, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt)
		if i != len(assets)-1 {
			builder.WriteByte(',')
		}
	}

	// insert into %s (%s) values (?,?,?,?),(?,?,?,?)
	query := fmt.Sprintf(sqlBatchInsert, builder.String())
	_, err := sess.ExecCtx(ctx, query, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *Repo) BatchInsert(ctx context.Context, assets []*Model) error {
	return r.batchInsert(ctx, r.db, assets)
}
func (r *Repo) BatchInsertTx(ctx context.Context, tx sqlx.Session, assets []*Model) error {
	return r.batchInsert(ctx, tx, assets)
}

func (r *Repo) FindByNoteIds(ctx context.Context, noteIds []uint64) ([]*Model, error) {
	if len(noteIds) == 0 {
		return []*Model{}, nil
	}
	query := fmt.Sprintf(sqlFindByNoteIds, uslices.JoinInts(noteIds))
	res := make([]*Model, 0)
	err := r.db.QueryRowsCtx(ctx, &res, query)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}
