package data

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
)

// 同步/异步写入数据接口
// 同步为数据库写入，异步为MQ写入
type Proxy interface {
	AddReply(ctx context.Context, data *dao.Comment) error
	DelReply(ctx context.Context, rid int64, reply *dao.Comment) error
	LikeReply(ctx context.Context, rid int64, uid int64) error
	UnLikeReply(ctx context.Context, rid int64, uid int64) error
	DisLikeReply(ctx context.Context, rid int64, uid int64) error
	UnDisLikeReply(ctx context.Context, rid int64, uid int64) error
	PinReply(ctx context.Context, oid, rid int64) error
	UnPinReply(ctx context.Context, oid, rid int64) error
}
