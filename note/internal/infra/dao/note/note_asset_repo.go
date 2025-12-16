package note

import (
	"context"
	"fmt"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	"github.com/huandu/go-sqlbuilder"
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

const (
	assetTable = "note_asset"
)

type AssetPO struct {
	Id        int64           `db:"id"`
	AssetKey  string          `db:"asset_key"`  // 资源key 包含bucket name
	AssetType model.AssetType `db:"asset_type"` // 资源类型
	NoteId    int64           `db:"note_id"`    // 所属笔记id
	CreateAt  int64           `db:"create_at"`  // 创建时间
	AssetMeta []byte          `db:"asset_meta"` // 资源的元数据 存储格式为一个json字符串
}

func (p *AssetPO) Values() []any {
	return []any{
		p.Id,
		p.AssetKey,
		p.AssetType,
		p.NoteId,
		p.CreateAt,
		p.AssetMeta,
	}
}

var (
	assetFields    = xsql.GetFieldSlice(&AssetPO{})
	assetInsFields = xsql.GetFieldSlice(&AssetPO{}, "id")
)

func (p *AssetPO) InsertValues() []any {
	var assetMeta []byte = p.AssetMeta
	if assetMeta == nil {
		assetMeta = []byte{}
	}
	return []any{
		p.AssetKey,
		p.AssetType,
		p.NoteId,
		p.CreateAt,
		assetMeta,
	}
}

func (r *NoteAssetRepo) FindOne(ctx context.Context, id int64) (*AssetPO, error) {
	model := new(AssetPO)
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(assetFields...).From(assetTable).Where(sb.Equal("id", id))
	sql, args := sb.Build()
	err := r.db.QueryRowCtx(ctx, model, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return model, nil
}

// 单条插入
func (r *NoteAssetRepo) Insert(ctx context.Context, asset *AssetPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(assetTable)
	ib.Cols(assetInsFields...)
	ib.Values(asset.Values()...)
	sql, args := ib.Build()

	_, err := r.db.ExecCtx(ctx, sql, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}

// 批量插入
func (r *NoteAssetRepo) BatchInsert(ctx context.Context, assets []*AssetPO) error {
	if len(assets) == 0 {
		return nil
	}

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(assetTable)
	ib.Cols(assetInsFields...)
	for _, data := range assets {
		ib.Values(data.InsertValues()...)
	}
	sql, args := ib.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) FindByNoteIdForUpdate(ctx context.Context, noteId int64, assetType model.AssetType) ([]*AssetPO, error) {
	res := make([]*AssetPO, 0)
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(assetFields...).
		From(assetTable).
		Where(
			sb.Equal("note_id", noteId),
			sb.Equal("asset_type", assetType),
		).ForUpdate()
	sql, args := sb.Build()
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

func (r *NoteAssetRepo) DeleteByNoteId(ctx context.Context, noteId int64) error {
	sb := sqlbuilder.NewDeleteBuilder()
	sb.DeleteFrom(assetTable)
	sb.Where(sb.Equal("note_id", noteId))
	sql, args := sb.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) ExcludeDeleteByNoteId(ctx context.Context, noteId int64, assetKeys []string, assetType model.AssetType) error {
	sb := sqlbuilder.NewDeleteBuilder()
	sb.DeleteFrom(assetTable)
	sb.Where(sb.Equal("note_id", noteId), sb.Equal("asset_type", assetType))
	if len(assetKeys) != 0 {
		sb.Where(sb.NotIn("asset_key", xslice.Any(assetKeys)...))
	}

	sql, args := sb.Build()
	_, err := r.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

func (r *NoteAssetRepo) FindByNoteIds(ctx context.Context, noteIds []int64) ([]*AssetPO, error) {
	if len(noteIds) == 0 {
		return []*AssetPO{}, nil
	}
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(assetFields...).From(assetTable).Where(sb.In("note_id", xslice.Any(noteIds)...))
	sql, args := sb.Build()
	res := make([]*AssetPO, 0)
	err := r.db.QueryRowsCtx(ctx, &res, sql, args...)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}
	return res, nil
}

// 批量更新asset_meta
func (r *NoteAssetRepo) BatchUpdateAssetMeta(ctx context.Context, noteId int64, metas map[string][]byte) error {
	if len(metas) == 0 {
		return nil
	}

	//	UPDATE note_asset SET asset_meta = CASE asset_key
	//	  WHEN ? THEN ?
	//	  WHEN ? THEN ?
	//	  ELSE asset_meta
	//	END
	//	WHERE note_id = ? AND asset_key IN (?, ?)
	var (
		whens        []string
		args         []any
		placeholders []string
		keys         []string // 保证顺序一致
	)

	for assetKey, assetMeta := range metas {
		whens = append(whens, "WHEN ? THEN ?")
		args = append(args, assetKey, assetMeta)
		placeholders = append(placeholders, "?")
		keys = append(keys, assetKey)
	}

	// 添加 note_id 和 asset_keys
	args = append(args, noteId)
	for _, key := range keys {
		args = append(args, key)
	}

	sql := fmt.Sprintf(
		"UPDATE %s SET asset_meta = CASE asset_key %s ELSE asset_meta END WHERE note_id = ? AND asset_key IN (%s)",
		assetTable,
		strings.Join(whens, " "),
		strings.Join(placeholders, ", "),
	)

	_, err := r.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}
