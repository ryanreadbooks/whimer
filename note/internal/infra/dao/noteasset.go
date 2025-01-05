package dao

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
	sqlFindAllById    = `SELECT id,asset_key,asset_type,note_id,create_at,asset_meta FROM note_asset WHERE id=?`
	sqlInsert         = `INSERT INTO note_asset(asset_key,asset_type,note_id,create_at,asset_meta) VALUES (?,?,?,?,?)`
	sqlDeleteByNoteId = `DELETE FROM note_asset WHERE note_id=?`
	sqlBatchInsert    = `INSERT INTO note_asset(asset_key,asset_type,note_id,create_at,asset_meta) VALUES %s`
	sqlFindByNoteIds  = `SELECT id,asset_key,asset_type,note_id,create_at,asset_meta FROM note_asset WHERE note_id in (%s)`

	sqlFindImageByNoteId          = `SELECT id,asset_key,asset_type,note_id,create_at,asset_meta FROM note_asset WHERE note_id=? AND asset_type=1 FOR UPDATE`
	sqlExcludeDeleteImageByNoteId = `DELETE FROM note_asset WHERE note_id=? AND asset_type=1`
	sqlFindImageByKey             = `SELECT id,asset_key,asset_type,note_id,create_at,asset_meta FROM note_asset WHERE asset_key=? AND asset_type=1 LIMIT 1`
)

type NoteAssetDao struct {
	db sqlx.SqlConn
}

func NewNoteAssetDao(db sqlx.SqlConn) *NoteAssetDao {
	return &NoteAssetDao{
		db: db,
	}
}

type NoteAsset struct {
	Id        uint64 `db:"id"`
	AssetKey  string `db:"asset_key"`  // 资源key 包含bucket name
	AssetType int8   `db:"asset_type"` // 资源类型
	NoteId    uint64 `db:"note_id"`    // 所属笔记id
	CreateAt  int64  `db:"create_at"`  // 创建时间
	AssetMeta string `db:"asset_meta"` // 资源的元数据 存储格式为一个json字符串
}

func (r *NoteAssetDao) FindOne(ctx context.Context, id uint64) (*NoteAsset, error) {
	model := new(NoteAsset)
	err := r.db.QueryRowCtx(ctx, model, sqlFindAllById, id)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return model, nil
}

func (r *NoteAssetDao) insert(ctx context.Context, sess sqlx.Session, asset *NoteAsset) error {
	now := time.Now().Unix()
	_, err := sess.ExecCtx(ctx, sqlInsert,
		asset.AssetKey,
		asset.AssetType,
		asset.NoteId,
		now,
		asset.AssetMeta)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetDao) Insert(ctx context.Context, asset *NoteAsset) error {
	return r.insert(ctx, r.db, asset)
}

func (r *NoteAssetDao) InsertTx(ctx context.Context, tx sqlx.Session, asset *NoteAsset) error {
	return r.insert(ctx, tx, asset)
}

func (r *NoteAssetDao) findImageByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64) ([]*NoteAsset, error) {
	res := make([]*NoteAsset, 0)
	err := sess.QueryRowsCtx(ctx, &res, sqlFindImageByNoteId, noteId)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteAssetDao) FindImageByNoteId(ctx context.Context, noteId uint64) ([]*NoteAsset, error) {
	return r.findImageByNoteId(ctx, r.db, noteId)
}

func (r *NoteAssetDao) FindImageByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64) ([]*NoteAsset, error) {
	return r.findImageByNoteId(ctx, tx, noteId)
}

func (r *NoteAssetDao) deleteByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64) error {
	_, err := sess.ExecCtx(ctx, sqlDeleteByNoteId, noteId)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetDao) DeleteByNoteId(ctx context.Context, noteId uint64) error {
	return r.deleteByNoteId(ctx, r.db, noteId)
}

func (r *NoteAssetDao) DeleteByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64) error {
	return r.deleteByNoteId(ctx, tx, noteId)
}

func (r *NoteAssetDao) excludeDeleteImageByNoteId(ctx context.Context, sess sqlx.Session, noteId uint64, assetKeys []string) error {
	var alen = len(assetKeys)
	var args []any = make([]any, 0, alen)
	args = append(args, noteId)

	query := sqlExcludeDeleteImageByNoteId
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
func (r *NoteAssetDao) ExcludeDeleteImageByNoteId(ctx context.Context, noteId uint64, assetKeys []string) error {
	return r.excludeDeleteImageByNoteId(ctx, r.db, noteId, assetKeys)
}

func (r *NoteAssetDao) ExcludeDeleteImageByNoteIdTx(ctx context.Context, tx sqlx.Session, noteId uint64, assetKeys []string) error {
	return r.excludeDeleteImageByNoteId(ctx, tx, noteId, assetKeys)
}

func (r *NoteAssetDao) batchInsert(ctx context.Context, sess sqlx.Session, assets []*NoteAsset) error {
	if len(assets) == 0 {
		return nil
	}

	tmpl := "(?, ?, ?, ?, ?)"
	var builder strings.Builder
	var args []any = make([]any, 0, len(assets)*4)
	for i, data := range assets {
		builder.WriteString(tmpl)
		args = append(args, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt, data.AssetMeta)
		if i != len(assets)-1 {
			builder.WriteByte(',')
		}
	}

	// insert into %s (%s) values (?,?,?,?,?),(?,?,?,?,?)
	query := fmt.Sprintf(sqlBatchInsert, builder.String())
	_, err := sess.ExecCtx(ctx, query, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetDao) BatchInsert(ctx context.Context, assets []*NoteAsset) error {
	return r.batchInsert(ctx, r.db, assets)
}
func (r *NoteAssetDao) BatchInsertTx(ctx context.Context, tx sqlx.Session, assets []*NoteAsset) error {
	return r.batchInsert(ctx, tx, assets)
}

func (r *NoteAssetDao) FindByNoteIds(ctx context.Context, noteIds []uint64) ([]*NoteAsset, error) {
	if len(noteIds) == 0 {
		return []*NoteAsset{}, nil
	}
	query := fmt.Sprintf(sqlFindByNoteIds, uslices.JoinInts(noteIds))
	res := make([]*NoteAsset, 0)
	err := r.db.QueryRowsCtx(ctx, &res, query)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteAssetDao) FindImageAssetByKey(ctx context.Context, assetKey string) (*NoteAsset, error) {
	var asset NoteAsset
	err := r.db.QueryRowCtx(ctx, &asset, sqlFindImageByKey, assetKey)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	return &asset, nil
}
