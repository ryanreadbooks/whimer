package queue

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/misc/utils"

	"github.com/zeromicro/go-queue/kq"
)

const (
	ActAddReply = 1 + iota
	ActDelReply
	ActLikeReply
	ActDislikeReply
)

const (
	ActionUndo = 0
	ActionDo   = 1
)

const (
	LikeType    = 0
	DisLikeType = 1
)

type (
	// 发表评论所需数据
	AddReplyData comm.Model

	// 删除评论所需数据
	DelReplyData struct {
		ReplyId uint64 `json:"reply_id"`
	}

	LikeReplyData struct {
		ReplyId uint64 `json:"reply_id"`
		Action  int    `json:"action"`
		Type    int    `json:"type"`
	}
)

// 放进消息队列中的数据
type Data struct {
	Action        int            `json:"action"`
	AddReplyData  *AddReplyData  `json:"add_reply_data,omitempty"`
	DelReplyData  *DelReplyData  `json:"del_reply_data,omitempty"`
	LikeReplyData *LikeReplyData `json:"like_reply_data,omitempty"`
}
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
		key = addDelKey
	default:
		key = otherKey
	}

	return b.pusher.PushWithKey(ctx, key, utils.Bytes2String(cd))
}

func (b *Bus) AddReply(ctx context.Context, data *comm.Model) error {
	return b.pushReplyAct(ctx, &Data{
		Action:       ActAddReply,
		AddReplyData: (*AddReplyData)(data),
	})
}

func (b *Bus) DelReply(ctx context.Context, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDelReply,
		DelReplyData: &DelReplyData{
			ReplyId: rid,
		},
	})
}

func (b *Bus) LikeReply(ctx context.Context, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActLikeReply,
		LikeReplyData: &LikeReplyData{
			ReplyId: rid,
			Action:  ActionDo,
			Type:    LikeType,
		},
	})
}

func (b *Bus) UnLikeReply(ctx context.Context, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActLikeReply,
		LikeReplyData: &LikeReplyData{
			ReplyId: rid,
			Action:  ActionUndo,
			Type:    LikeType,
		},
	})
}

func (b *Bus) DisLikeReply(ctx context.Context, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDislikeReply,
		LikeReplyData: &LikeReplyData{
			ReplyId: rid,
			Action:  ActionDo,
			Type:    DisLikeType,
		},
	})
}

func (b *Bus) UndisLikeReply(ctx context.Context, rid uint64) error {
	return b.pushReplyAct(ctx, &Data{
		Action: ActDislikeReply,
		LikeReplyData: &LikeReplyData{
			ReplyId: rid,
			Action:  ActionUndo,
			Type:    DisLikeType,
		},
	})
}
