package note

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	uslices "github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/model"
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
)

// NoteAssetRepo 笔记资源数据库仓储 - 纯数据库操作
type NoteAssetRepo struct {
	db *xsql.DB
}

func NewNoteAssetRepo(db *xsql.DB) *NoteAssetRepo {
	return &NoteAssetRepo{
		db: db,
	}
}

type AssetPO struct {
	Id        int64           `db:"id"`
	AssetKey  string          `db:"asset_key"`  // 资源key 包含bucket name
	AssetType model.AssetType `db:"asset_type"` // 资源类型
	NoteId    int64           `db:"note_id"`    // 所属笔记id
	CreateAt  int64           `db:"create_at"`  // 创建时间
	AssetMeta []byte          `db:"asset_meta"` // 资源的元数据 存储格式为一个json字符串
}

func (r *NoteAssetRepo) FindOne(ctx context.Context, id int64) (*AssetPO, error) {
	model := new(AssetPO)
	err := r.db.QueryRowCtx(ctx, model, sqlFindAllById, id)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return model, nil
}

func (r *NoteAssetRepo) insert(ctx context.Context, asset *AssetPO) error {
	var (
		assetMeta []byte = asset.AssetMeta
	)
	if assetMeta == nil {
		assetMeta = []byte{}
	}
	now := time.Now().Unix()
	_, err := r.db.ExecCtx(ctx, sqlInsert,
		asset.AssetKey,
		asset.AssetType,
		asset.NoteId,
		now,
		assetMeta)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) Insert(ctx context.Context, asset *AssetPO) error {
	return r.insert(ctx, asset)
}

func (r *NoteAssetRepo) findImageByNoteId(ctx context.Context, noteId int64) ([]*AssetPO, error) {
	res := make([]*AssetPO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sqlFindImageByNoteId, noteId)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteAssetRepo) FindImageByNoteId(ctx context.Context, noteId int64) ([]*AssetPO, error) {
	return r.findImageByNoteId(ctx, noteId)
}

func (r *NoteAssetRepo) deleteByNoteId(ctx context.Context, noteId int64) error {
	_, err := r.db.ExecCtx(ctx, sqlDeleteByNoteId, noteId)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) DeleteByNoteId(ctx context.Context, noteId int64) error {
	return r.deleteByNoteId(ctx, noteId)
}

func (r *NoteAssetRepo) excludeDeleteImageByNoteId(ctx context.Context, noteId int64, assetKeys []string) error {
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
		query += fmt.Sprintf(" AND `asset_key` NOT IN (%s)", tmpl)
	}
	_, err := r.db.ExecCtx(ctx, query, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}
func (r *NoteAssetRepo) ExcludeDeleteImageByNoteId(ctx context.Context, noteId int64, assetKeys []string) error {
	return r.excludeDeleteImageByNoteId(ctx, noteId, assetKeys)
}

func (r *NoteAssetRepo) batchInsert(ctx context.Context, assets []*AssetPO) error {
	if len(assets) == 0 {
		return nil
	}

	tmpl := "(?, ?, ?, ?, ?)"
	var builder strings.Builder
	var args []any = make([]any, 0, len(assets)*4)
	for i, data := range assets {
		builder.WriteString(tmpl)
		var assetMeta []byte = data.AssetMeta
		if assetMeta == nil {
			assetMeta = []byte{}
		}
		args = append(args, data.AssetKey, data.AssetType, data.NoteId, data.CreateAt, assetMeta)
		if i != len(assets)-1 {
			builder.WriteByte(',')
		}
	}

	// insert into %s (%s) values (?,?,?,?,?),(?,?,?,?,?)
	query := fmt.Sprintf(sqlBatchInsert, builder.String())
	_, err := r.db.ExecCtx(ctx, query, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) BatchInsert(ctx context.Context, assets []*AssetPO) error {
	return r.batchInsert(ctx, assets)
}
func (r *NoteAssetRepo) FindByNoteIds(ctx context.Context, noteIds []int64) ([]*AssetPO, error) {
	if len(noteIds) == 0 {
		return []*AssetPO{}, nil
	}
	query := fmt.Sprintf(sqlFindByNoteIds, uslices.JoinInts(noteIds))
	res := make([]*AssetPO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, query)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}
