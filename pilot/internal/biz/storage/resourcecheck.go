package storage

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

const (
	objectStatusTagKey    = "x-status"
	objectActiveStatusTag = "active"
)

const (
	v7ErrCodeNoSuchKey    = "NoSuchKey"
	v7ErrCodeNoSuchBucket = "NoSuchBucket"
)

var (
	tagMap       = map[string]string{objectStatusTagKey: objectActiveStatusTag}
	objectTag, _ = tags.NewTags(tagMap, true)
)

var (
	ErrResourceNotFound = fmt.Errorf("resource not found")
)

func isV7ErrNotFound(err error) (newErr error, ok bool) {
	var v7Err minio.ErrorResponse
	if errors.As(err, &v7Err) {
		if v7Err.Code == v7ErrCodeNoSuchKey || v7Err.Code == v7ErrCodeNoSuchBucket {
			return ErrResourceNotFound, true
		}
	}

	return err, false
}

// 检查oss中资源是否存在
func (b *Biz) CheckResourceExist(ctx context.Context, bucket, key string) (bool, error) {
	resp, err := b.ossClient.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return false, newErr
		}
		return false, xerror.Wrapf(err, "stat object failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	if resp.Key == key && !resp.IsDeleteMarker {
		return true, nil
	}

	return false, nil
}

// 检查oss中资源是否存在
func (b *Biz) BatchCheckResourceExist(ctx context.Context, bucket string, keys []string, strict bool) (map[string]bool, error) {
	// gopool concurrent
	res := make(map[string]bool, len(keys))
	for _, key := range keys {
		ok, err := b.CheckResourceExist(ctx, bucket, key)
		if err != nil && strict {
			return res, err
		}
		if ok {
			res[key] = true
		} else {
			res[key] = false
		}
	}

	return res, nil
}

func (b *Biz) MarkResourceActive(ctx context.Context, bucket, key string) error {
	err := b.ossClient.PutObjectTagging(ctx, bucket, key, objectTag, minio.PutObjectTaggingOptions{})
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return newErr
		}
		return xerror.Wrapf(err, "put object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	return nil
}

func (b *Biz) BatchMarkResourceActive(ctx context.Context, bucket string, keys []string, strict bool) error {
	for _, key := range keys {
		err := b.MarkResourceActive(ctx, bucket, key)
		if err != nil && strict {
			return err
		}
		if err != nil {
			xlog.Msg("batch mark resource active err").Err(err).Extras("bucket", bucket, "keys", keys).Errorx(ctx)
		}
	}

	return nil
}

func (b *Biz) MarkResourceInactive(ctx context.Context, bucket, key string) error {
	curTags, err := b.ossClient.GetObjectTagging(ctx, bucket, key, minio.GetObjectTaggingOptions{})
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return newErr
		}
		return xerror.Wrapf(err, "get object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	curMap := curTags.ToMap()
	delete(curMap, objectStatusTagKey)

	t, _ := tags.MapToObjectTags(curMap)
	err = b.ossClient.PutObjectTagging(ctx, bucket, key, t, minio.PutObjectTaggingOptions{})
	if err != nil {
		return xerror.Wrapf(err, "put object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	return nil
}

// 获取资源的numBytes大小数据
//
// numBytes<=0表示获取全部
func (b *Biz) GetResourceBytes(ctx context.Context, bucket, key string, numBytes int32) ([]byte, int64, error) {
	getOpt := minio.GetObjectOptions{}
	if numBytes > 0 {
		getOpt.Set("Range", fmt.Sprintf("bytes=0-%d", numBytes-1))
	}

	resp, err := b.ossClient.GetObject(ctx, bucket, key, getOpt)
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return nil, 0, newErr
		}
		return nil, 0, xerror.Wrapf(err, "get object failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	content, err := io.ReadAll(resp)
	if err != nil {
		return nil, 0, xerror.Wrapf(xerror.ErrInternal.Msg(err.Error()), "io readall failed").
			WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}
	defer resp.Close()

	stat, err := resp.Stat()
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "stat resp failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}
	return content, stat.Size, nil
}
