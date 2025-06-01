package bus

import "github.com/ryanreadbooks/whimer/comment/internal/infra/dao"

const (
	ActAddReply = 1 + iota
	ActDelReply
	ActLikeReply
	ActDislikeReply
	ActPinReply
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
	AddReplyData dao.Comment

	// 删除评论所需数据
	DelReplyData struct {
		ReplyId uint64       `json:"reply_id"`
		Reply   *dao.Comment `json:"reply"`
	}

	BinaryReplyData struct {
		Uid     int64  `json:"uid"`
		ReplyId uint64 `json:"reply_id"`
		Action  int    `json:"action"` // do or undo
		Type    int    `json:"type"`   // like or dislike
	}

	PinReplyData struct {
		ReplyId uint64 `json:"reply_id"`
		Action  int    `json:"action"` // do or undo
		Oid     uint64 `json:"oid"`
	}
)

// 放进消息队列中的数据
type Data struct {
	Action        int              `json:"action"`
	AddReplyData  *AddReplyData    `json:"add_reply_data,omitempty"`
	DelReplyData  *DelReplyData    `json:"del_reply_data,omitempty"`
	LikeReplyData *BinaryReplyData `json:"like_reply_data,omitempty"`
	PinReplyData  *PinReplyData    `json:"pin_reply_data,omitempty"`
}
