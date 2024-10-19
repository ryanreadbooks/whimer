package svc

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
)

// 同步/异步写入数据接口
type IDataProxy interface {
	AddReply(ctx context.Context, data *comm.Model) error
	DelReply(ctx context.Context, rid uint64, reply *comm.Model) error
	LikeReply(ctx context.Context, rid, uid uint64) error
	UnLikeReply(ctx context.Context, rid, uid uint64) error
	DisLikeReply(ctx context.Context, rid, uid uint64) error
	UnDisLikeReply(ctx context.Context, rid, uid uint64) error
	PinReply(ctx context.Context, oid, rid uint64) error
	UnPinReply(ctx context.Context, oid, rid uint64) error
}
