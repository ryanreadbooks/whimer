package dao

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type CommentAssetDao struct {
	db    *xsql.DB
	cache *CommentAssetCache
}

func NewCommentAssetDao(db *xsql.DB, cache *redis.Redis) *CommentAssetDao {
	return &CommentAssetDao{
		db:    db,
		cache: NewCommentAssetCache(cache),
	}
}

var (
	_assetInst = &CommentAsset{}

	assetFields    = xsql.GetFields(_assetInst)
	_, insAssetQst = xsql.GetFields2WithSkip(_assetInst, "id") // for insert

	sqlGetAssetByCommentId       = fmt.Sprintf("SELECT %s FROM comment_asset WHERE comment_id=?", assetFields)
	sqlBatchGetAssetByCommentIds = fmt.Sprintf("SELECT %s FROM comment_asset WHERE comment_id IN (%%s)", assetFields)
	sqlBatchInsertAsset          = "INSERT INTO comment_asset(comment_id,type,store_key,metadata,ctime) VALUES %s"
	sqlDelByCommentId            = "DELETE FROM comment_asset WHERE comment_id=?"
	sqlBatchDelByCommentId       = "DELETE FROM comment_asset WHERE comment_id IN (%s)"

	// 先查comment再删comment_asset
	sqlBatchDelAssetsBySelectedCommentId = "DELETE FROM comment_asset WHERE comment_id IN (SELECT id FROM comment WHERE root=?) OR comment_id=?"
)

// 批量插入
func (d *CommentAssetDao) BatchInsert(ctx context.Context, assets []*CommentAsset) error {
	if len(assets) == 0 {
		return nil
	}

	now := time.Now().Unix()
	tmpl := "(" + insAssetQst + ")"

	// insert into %s(%s) VALUES (?,?...),(?,?,...)
	var sql = fmt.Sprintf(sqlBatchInsertAsset, xslice.JoinStrings(xslice.Repeat(tmpl, len(assets))))
	var args = make([]any, 0, len(assets)*5)
	for _, asset := range assets {

		ctime := asset.Ctime
		if ctime == 0 {
			ctime = now
		}
		metadata := asset.Metadata
		if metadata == nil {
			metadata = json.RawMessage{}
		}

		args = append(args, asset.CommentId, asset.Type, asset.StoreKey, metadata, ctime)
	}

	_, err := d.db.ExecCtx(ctx, sql, args...)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (d *CommentAssetDao) GetByCommentId(ctx context.Context, cid int64) ([]*CommentAsset, error) {
	// 先从缓存获取
	cacheAssets, err := d.cache.GetByCommentId(ctx, cid)
	if err == nil && len(cacheAssets) > 0 {
		return cacheAssets, nil
	}

	var ret []*CommentAsset
	err = d.db.QueryRowsCtx(ctx, &ret, sqlGetAssetByCommentId, cid)
	if err != nil {
		return nil, xerror.Wrap(xsql.ConvertError(err))
	}

	if len(ret) > 0 {
		d.cache.SetByCommentId(ctx, cid, ret)
	}

	return ret, nil
}

func (d *CommentAssetDao) BatchGetByCommentIds(ctx context.Context, cids []int64) (map[int64][]*CommentAsset, error) {
	ret := make(map[int64][]*CommentAsset)
	if len(cids) == 0 {
		return ret, nil
	}

	cacheAssets, _ := d.cache.BatchGetByCommentIds(ctx, cids)
	missingCids := make([]int64, 0, len(cids))
	for _, cid := range cids {
		if assets, exists := cacheAssets[cid]; !exists || len(assets) == 0 {
			missingCids = append(missingCids, cid)
		}
	}

	if len(missingCids) == 0 {
		return cacheAssets, nil
	}

	// 从数据库获取未命中的
	dbAssets := make(map[int64][]*CommentAsset, len(missingCids))
	err := xslice.BatchExec(missingCids, 100, func(start, end int) error {
		targets := missingCids[start:end]
		var sql = fmt.Sprintf(sqlBatchGetAssetByCommentIds, xslice.JoinStrings(xslice.Repeat("?", len(targets))))
		var res []*CommentAsset
		args := xslice.Any(targets)
		err := d.db.QueryRowsCtx(ctx, &res, sql, args...)
		if err != nil {
			return xerror.Wrap(xsql.ConvertError(err))
		}

		for _, item := range res {
			dbAssets[item.CommentId] = append(dbAssets[item.CommentId], item)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// 合并结果
	for cid, cacheAsset := range cacheAssets {
		ret[cid] = append(ret[cid], cacheAsset...)
	}
	for cid, dbAsset := range dbAssets {
		ret[cid] = append(ret[cid], dbAsset...)
	}

	tmpRet := make(map[int64][]*CommentAsset, len(ret))
	for cid, assets := range ret {
		deDupAssets := xslice.UniqF(assets, func(v *CommentAsset) int64 {
			return v.Id
		})
		tmpRet[cid] = deDupAssets
	}

	if len(dbAssets) > 0 {
		d.cache.BatchSetByCommentIdsAsync(ctx, dbAssets)
	}

	return tmpRet, nil
}

func (d *CommentAssetDao) DeleteByCommentId(ctx context.Context, cid int64) error {
	_, err := d.db.ExecCtx(ctx, sqlDelByCommentId, cid)
	d.cache.DeleteByCommentIdAsync(ctx, cid)

	return xerror.Wrap(xsql.ConvertError(err))
}

func (d *CommentAssetDao) BatchDeleteByCommentId(ctx context.Context, cids []int64) error {
	if len(cids) == 0 {
		return nil
	}

	// delete from comment_asset where comment_id in (?,?,...,?)
	var sql = fmt.Sprintf(sqlBatchDelByCommentId, xslice.JoinStrings(xslice.Repeat("?", len(cids))))
	err := xslice.BatchExec(cids, 200, func(start, end int) error {
		targets := xslice.Any(cids[start:end])
		_, err := d.db.ExecCtx(ctx, sql, targets...)
		return xerror.Wrap(xsql.ConvertError(err))
	})

	d.cache.BatchDeleteByCommentIdsAsync(ctx, cids)

	return err
}

// 删除根评论为root的子评论的所有asset, 并且一并删除root的asset
func (d *CommentAssetDao) BatchDeleteByRoot(ctx context.Context, root int64) error {
	_, err := d.db.ExecCtx(ctx, sqlBatchDelAssetsBySelectedCommentId, root, root)
	// 不主动处理缓存 等自动过期即可
	return xerror.Wrap(xsql.ConvertError(err))
}
