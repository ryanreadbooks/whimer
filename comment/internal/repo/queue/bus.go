package queue

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
)

type Bus struct {
	topic string
}

func New(c *config.Config) *Bus {
	return &Bus{}
}

const (
	ActAddReply = 1 + iota
	ActDelReply
)

type (
	// 发表评论所需数据
	AddReplyData comm.Model

	// 删除评论所需数据
	DelReplyData struct {
		ReplyId uint64 `json:"reply_id"`
	}
)

// 放进消息队列中的数据
type Data struct {
	Action       int           `json:"action"`
	AddReplyData *AddReplyData `json:"add_reply_data,omitempty"`
	DelReplyData *DelReplyData `json:"del_reply_data,omitempty"`
}

func (b *Bus) pushReplyAct(ctx context.Context, data *Data) error {

	return nil
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
