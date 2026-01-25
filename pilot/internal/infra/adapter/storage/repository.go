package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	xhttputil "github.com/ryanreadbooks/whimer/misc/xhttp/util"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	domainstorage "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
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

var ErrObjectNotFound = fmt.Errorf("object not found")

var _ domainstorage.Repository = (*OssRepositoryImpl)(nil)

type OssRepositoryImpl struct {
	uploaders        *Uploaders
	ossClient        *minio.Client
	displayOssClient *minio.Client
}

func NewOssRepositoryImpl(
	uploaders *Uploaders,
	ossClient *minio.Client,
	displayOssClient *minio.Client,
) *OssRepositoryImpl {
	return &OssRepositoryImpl{
		uploaders:        uploaders,
		ossClient:        ossClient,
		displayOssClient: displayOssClient,
	}
}

// ==================== ObjectChecker ====================

func (r *OssRepositoryImpl) CheckFileIdValid(objType vo.ObjectType, fileId string) error {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return err
	}
	return uploader.CheckFileIdValid(fileId)
}

func (r *OssRepositoryImpl) CheckObjectExist(
	ctx context.Context, objType vo.ObjectType, fileId string,
) (bool, error) {
	bucket, key, err := r.SeperateObject(objType, fileId)
	if err != nil {
		return false, err
	}
	return r.checkObjectExistByKey(ctx, bucket, key)
}

func (r *OssRepositoryImpl) checkObjectExistByKey(ctx context.Context, bucket, key string) (bool, error) {
	resp, err := r.ossClient.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
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

func (r *OssRepositoryImpl) CheckAndMarkObjects(
	ctx context.Context, objType vo.ObjectType, objects []vo.ObjectInfo,
) error {
	if len(objects) == 0 {
		return nil
	}

	fileIds := make([]string, 0, len(objects))
	for _, obj := range objects {
		if obj.FileId == "" {
			continue
		}
		fileIds = append(fileIds, obj.FileId)
	}

	if len(fileIds) == 0 {
		return nil
	}

	// 检查 fileId 格式
	for _, fileId := range fileIds {
		if err := r.CheckFileIdValid(objType, fileId); err != nil {
			return err
		}
	}

	bucket, keys, err := r.collectObjectKeys(objType, fileIds)
	if err != nil {
		return err
	}

	// 批量检查对象是否存在
	for _, key := range keys {
		exists, err := r.checkObjectExistByKey(ctx, bucket, key)
		if err != nil {
			return err
		}
		if !exists {
			return ErrObjectNotFound
		}
	}

	// 检查对象内容
	if err = r.checkObjectsContent(ctx, objType, bucket, keys); err != nil {
		return err
	}

	// 标记对象为激活状态
	r.batchMarkObjectActive(ctx, bucket, keys, false)
	return nil
}

func (r *OssRepositoryImpl) UnmarkObjects(
	ctx context.Context, objType vo.ObjectType, objects []vo.ObjectInfo,
) error {
	if len(objects) == 0 {
		return nil
	}

	fileIds := make([]string, 0, len(objects))
	for _, obj := range objects {
		if obj.FileId == "" {
			continue
		}
		fileIds = append(fileIds, obj.FileId)
	}

	if len(fileIds) == 0 {
		return nil
	}

	bucket, keys, err := r.collectObjectKeys(objType, fileIds)
	if err != nil {
		return err
	}

	r.batchMarkObjectInactive(ctx, bucket, keys, false)
	return nil
}

func (r *OssRepositoryImpl) CheckObjectContent(
	ctx context.Context, objType vo.ObjectType, fileId string,
) error {
	bucket, key, err := r.SeperateObject(objType, fileId)
	if err != nil {
		return err
	}
	content, size, err := r.getObjectBytes(ctx, bucket, key, 32)
	if err != nil {
		return err
	}
	return objType.CheckContent(content, size)
}

// ==================== ObjectPathResolver ====================

func (r *OssRepositoryImpl) SeperateObject(
	objType vo.ObjectType, fileId string,
) (bucket, key string, err error) {
	return r.uploaders.SeperateObject(r.toUploadObjectType(objType), fileId)
}

func (r *OssRepositoryImpl) TrimBucketAndPrefix(objType vo.ObjectType, fileId string) string {
	return r.uploaders.TrimBucketAndPrefix(r.toUploadObjectType(objType), fileId)
}

func (r *OssRepositoryImpl) ResolveTargetPath(
	objType vo.ObjectType, fileId string,
) (targetPath string, err error) {
	rawKey := r.uploaders.TrimBucketAndPrefix(r.toUploadObjectType(objType), fileId)
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return "", err
	}
	meta := uploader.GetMetadata()
	targetPath = meta.Bucket + "/" + meta.PrefixSegment + "/" + rawKey
	return targetPath, nil
}

func (r *OssRepositoryImpl) GetBucket(objType vo.ObjectType) (string, error) {
	return r.uploaders.GetBucket(r.toUploadObjectType(objType))
}

func (r *OssRepositoryImpl) GetObjectMeta(objType vo.ObjectType) (*vo.ObjectMeta, error) {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return nil, err
	}
	meta := uploader.GetMetadata()
	return &vo.ObjectMeta{
		Bucket:        meta.Bucket,
		Prefix:        meta.Prefix,
		PrefixSegment: meta.PrefixSegment,
	}, nil
}

// ==================== UploadTicketProvider ====================

func (r *OssRepositoryImpl) GetUploadTicket(
	ctx context.Context, objType vo.ObjectType, count int32,
) (*vo.UploadTicket, error) {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return nil, err
	}

	tmpCreds, err := uploader.GetCredentials(ctx)
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader %s failed to get sts credentials", objType).WithCtx(ctx)
	}

	return &vo.UploadTicket{
		FileIds:      uploader.GetFileIds(count),
		Bucket:       uploader.GetBucket(),
		AccessKey:    tmpCreds.AccessKeyID,
		SecretKey:    tmpCreds.SecretAccessKey,
		SessionToken: tmpCreds.SessionToken,
		ExpireAt:     tmpCreds.Expiration.Unix(),
		UploadAddr:   xhttputil.FormatHost(uploader.GetUploadEndpoint(), false),
	}, nil
}

// TODO
//
// Should be deprecated in the future
func (r *OssRepositoryImpl) GetUploadTicketDeprecated(
	ctx context.Context, objType vo.ObjectType, count int32,
) (*vo.UploadTicketDeprecated, error) {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return nil, err
	}

	ticket, err := uploader.GenerateUploadTicket(count, "")
	if err != nil {
		return nil, xerror.Wrapf(err, "uploader generate ticket failed").WithCtx(ctx)
	}

	return &vo.UploadTicketDeprecated{
		FileIds:     ticket.FileIds,
		CurrentTime: ticket.CurrentTime,
		ExpireTime:  ticket.ExpireTime,
		UploadAddr:  ticket.UploadAddr,
		Token:       ticket.Token,
	}, nil
}

func (r *OssRepositoryImpl) GetPostPolicyTicket(
	ctx context.Context, objType vo.ObjectType, sha256 string, mimeType string,
) (*vo.PostPolicyTicket, error) {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return nil, err
	}

	ppResp, err := uploader.GetPostPolicy(ctx, &GetPostPolicyRequest{
		ContentType: mimeType,
		Sha256:      sha256,
	})
	if err != nil {
		return nil, err
	}

	return &vo.PostPolicyTicket{
		UploadAddr: ppResp.Url,
		FileId:     ppResp.Key,
		Form:       ppResp.Form,
	}, nil
}

func (r *OssRepositoryImpl) PresignGetUrl(
	ctx context.Context, objType vo.ObjectType, fileId string,
) (string, error) {
	uploader, err := r.uploaders.GetUploader(r.toUploadObjectType(objType))
	if err != nil {
		return "", err
	}
	bucket := uploader.GetBucket()
	_, rawKey, ok := uploader.keyGen.Unwrap(fileId)
	if !ok {
		return "", xerror.ErrArgs.Msg("资源格式错误")
	}

	presignedURL, err := r.displayOssClient.PresignedGetObject(ctx, bucket, rawKey, time.Hour, nil)
	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

// ==================== Internal helpers ====================

func (r *OssRepositoryImpl) toUploadObjectType(objType vo.ObjectType) uploadresource.Type {
	return uploadresource.Type(objType)
}

func (r *OssRepositoryImpl) collectObjectKeys(
	objType vo.ObjectType, fileIds []string,
) (bucket string, keys []string, err error) {
	keys = make([]string, 0, len(fileIds))
	for _, fileId := range fileIds {
		var key string
		bucket, key, err = r.SeperateObject(objType, fileId)
		if err != nil {
			return "", nil, err
		}
		keys = append(keys, key)
	}
	return bucket, keys, nil
}

func (r *OssRepositoryImpl) checkObjectsContent(
	ctx context.Context, objType vo.ObjectType, bucket string, keys []string,
) error {
	for _, key := range keys {
		content, size, err := r.getObjectBytes(ctx, bucket, key, 32)
		if err != nil {
			return err
		}
		if err = objType.CheckContent(content, size); err != nil {
			return err
		}
	}
	return nil
}

func (r *OssRepositoryImpl) getObjectBytes(
	ctx context.Context, bucket, key string, numBytes int32,
) ([]byte, int64, error) {
	getOpt := minio.GetObjectOptions{}
	if numBytes > 0 {
		getOpt.Set("Range", fmt.Sprintf("bytes=0-%d", numBytes-1))
	}

	resp, err := r.ossClient.GetObject(ctx, bucket, key, getOpt)
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return nil, 0, newErr
		}
		return nil, 0, xerror.Wrapf(err, "get object failed").
			WithExtras("bucket", bucket, "key", key).
			WithCtx(ctx)
	}
	defer resp.Close()

	content, err := io.ReadAll(resp)
	if err != nil {
		return nil, 0, xerror.Wrapf(xerror.ErrInternal.Msg(err.Error()), "io readall failed").
			WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	stat, err := resp.Stat()
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "stat resp failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}
	return content, stat.Size, nil
}

func (r *OssRepositoryImpl) batchMarkObjectActive(
	ctx context.Context, bucket string, keys []string, strict bool,
) error {
	for _, key := range keys {
		err := r.markObjectActive(ctx, bucket, key)
		if err != nil && strict {
			return err
		}
		if err != nil {
			xlog.Msg("batch mark object active err").Err(err).Extras("bucket", bucket, "keys", keys).Errorx(ctx)
		}
	}
	return nil
}

func (r *OssRepositoryImpl) markObjectActive(ctx context.Context, bucket, key string) error {
	err := r.ossClient.PutObjectTagging(ctx, bucket, key, objectTag, minio.PutObjectTaggingOptions{})
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return newErr
		}
		return xerror.Wrapf(err, "put object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}
	return nil
}

func (r *OssRepositoryImpl) batchMarkObjectInactive(
	ctx context.Context, bucket string, keys []string, strict bool,
) error {
	for _, key := range keys {
		err := r.markObjectInactive(ctx, bucket, key)
		if err != nil {
			if strict {
				return err
			}
			xlog.Msg("batch mark object inactive err").Err(err).Extras("bucket", bucket, "keys", keys).Errorx(ctx)
		}
	}
	return nil
}

func (r *OssRepositoryImpl) markObjectInactive(ctx context.Context, bucket, key string) error {
	curTags, err := r.ossClient.GetObjectTagging(ctx, bucket, key, minio.GetObjectTaggingOptions{})
	if err != nil {
		if newErr, ok := isV7ErrNotFound(err); ok {
			return newErr
		}
		return xerror.Wrapf(err, "get object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	curMap := curTags.ToMap()
	delete(curMap, objectStatusTagKey)

	t, _ := tags.MapToObjectTags(curMap)
	err = r.ossClient.PutObjectTagging(ctx, bucket, key, t, minio.PutObjectTaggingOptions{})
	if err != nil {
		return xerror.Wrapf(err, "put object tagging failed").WithExtras("bucket", bucket, "key", key).WithCtx(ctx)
	}

	return nil
}

func isV7ErrNotFound(err error) (newErr error, ok bool) {
	var v7Err minio.ErrorResponse
	if errors.As(err, &v7Err) {
		if v7Err.Code == v7ErrCodeNoSuchKey || v7Err.Code == v7ErrCodeNoSuchBucket {
			return ErrObjectNotFound, true
		}
	}
	return err, false
}
