package storage

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
)

// 检查对象是否存在并受支持
type ObjectChecker interface {
	CheckFileIdValid(objType vo.ObjectType, fileId string) error
	CheckObjectExist(ctx context.Context, objType vo.ObjectType, fileId string) (bool, error)
	CheckAndMarkObjects(ctx context.Context, objType vo.ObjectType, objects []vo.ObjectInfo) error
	UnmarkObjects(ctx context.Context, objType vo.ObjectType, objects []vo.ObjectInfo) error
	CheckObjectContent(ctx context.Context, objType vo.ObjectType, fileId string) error
}

// 解析对象路径
type ObjectPathResolver interface {
	SeperateObject(objType vo.ObjectType, fileId string) (bucket, key string, err error)
	TrimBucketAndPrefix(objType vo.ObjectType, fileId string) string
	ResolveTargetPath(objType vo.ObjectType, fileId string) (targetPath string, err error)
	GetBucket(objType vo.ObjectType) (string, error)
	GetObjectMeta(objType vo.ObjectType) (*vo.ObjectMeta, error)
}

// 上传凭证处理
type UploadTicketProvider interface {
	GetUploadTicketDeprecated(ctx context.Context, objType vo.ObjectType, count int32) (*vo.UploadTicketDeprecated, error)
	GetUploadTicket(ctx context.Context, objType vo.ObjectType, count int32) (*vo.UploadTicket, error)
	GetPostPolicyTicket(ctx context.Context, objType vo.ObjectType, sha256 string, mimeType string) (*vo.PostPolicyTicket, error)
	PresignGetUrl(ctx context.Context, objType vo.ObjectType, fileId string) (string, error)
}

// Repository 整合所有存储相关能力
type Repository interface {
	ObjectChecker
	ObjectPathResolver
	UploadTicketProvider
}
