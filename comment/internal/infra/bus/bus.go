package bus

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/utils"

	"github.com/zeromicro/go-queue/kq"
)

type Bus struct {
	topic  string
	pusher *kq.Pusher
}

func New(c *config.Config) *Bus {
	b := Bus{
		topic:  c.Kafka.Topic,
		pusher: kq.NewPusher(c.Kafka.Brokers, c.Kafka.Topic),
	}

	return &b
}

func (b *Bus) pushReplyAct(ctx context.Context, data *Data) error {
	cd, err := json.Marshal(data)
	if err != nil {
		return err
	}

	const (
		addDelKey = "reply_addel"
		otherKey  = "reply_otkey"
	)

	var key string
	switch data.Action {
	case ActAddReply, ActDelReply:
		key = addDelKey // 新增和删除方同一个partition中
	default:
		key = otherKey
	}

	return b.pusher.PushWithKey(ctx, key, utils.Bytes2String(cd))
}

func (b *Bus) AddReply(ctx context.Context, data *dao.Comment) error {
	return b.pushReplyAct(ctx, &Data{
		Action:       ActAddReply,
		AddReplyData: (*AddReplyData)(data),
	})
}

func (b *Bus) DelReply(ctx context.Context, rid uint64, reply *dao.Comment) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDelReply,
		DelReplyData: &DelReplyData{
			ReplyId: rid,
			Reply:   reply,
		},
	})
}

func (b *Bus) LikeReply(ctx context.Context, rid, uid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActLikeReply,
		LikeReplyData: &BinaryReplyData{
			Uid:     uid,
			ReplyId: rid,
			Action:  ActionDo,
			Type:    LikeType,
		},
	})
}

func (b *Bus) UnLikeReply(ctx context.Context, rid, uid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActLikeReply,
		LikeReplyData: &BinaryReplyData{
			Uid:     uid,
			ReplyId: rid,
			Action:  ActionUndo,
			Type:    LikeType,
		},
	})
}

func (b *Bus) DisLikeReply(ctx context.Context, rid, uid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDislikeReply,
		LikeReplyData: &BinaryReplyData{
			Uid:     uid,
			ReplyId: rid,
			Action:  ActionDo,
			Type:    DisLikeType,
		},
	})
}

func (b *Bus) UnDisLikeReply(ctx context.Context, rid, uid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDislikeReply,
		LikeReplyData: &BinaryReplyData{
			Uid:     uid,
			ReplyId: rid,
			Action:  ActionUndo,
			Type:    DisLikeType,
		},
	})
}

// 置顶评论
func (b *Bus) PinReply(ctx context.Context, oid, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActPinReply,
		PinReplyData: &PinReplyData{
			ReplyId: rid,
			Action:  ActionDo,
			Oid:     oid,
		},
	})
}

// 取消置顶
func (b *Bus) UnPinReply(ctx context.Context, oid, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActPinReply,
		PinReplyData: &PinReplyData{
			ReplyId: rid,
			Action:  ActionUndo,
			Oid:     oid,
		},
	})
}
