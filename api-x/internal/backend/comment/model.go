package comment

import (
	"github.com/ryanreadbooks/whimer/comment/sdk"
	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/passport/sdk/user"
)

type PubReq struct {
	ReplyType uint32 `json:"reply_type"`
	Oid       uint64 `json:"oid"`
	Content   string `json:"content"`
	RootId    uint64 `json:"root_id,omitempty,optional"`
	ParentId  uint64 `json:"parent_id,omitempty,optional"`
	ReplyUid  uint64 `json:"reply_uid"`
}

func (r *PubReq) AsPb() *sdk.AddReplyReq {
	return &sdk.AddReplyReq{
		ReplyType: r.ReplyType,
		Oid:       r.Oid,
		Content:   r.Content,
		RootId:    r.RootId,
		ParentId:  r.ParentId,
		ReplyUid:  r.ReplyUid,
	}
}

type PubRes struct {
	ReplyId uint64 `json:"reply_id"`
}

type GetCommentsReq struct {
	Oid    uint64 `form:"oid"`
	Cursor uint64 `form:"cursor,optional"`
	SortBy int    `form:"sort_by,optional"`
}

func (r *GetCommentsReq) AsPb() *sdk.PageGetReplyReq {
	return &sdk.PageGetReplyReq{
		Oid:    r.Oid,
		Cursor: r.Cursor,
		SortBy: sdk.SortType(r.SortBy),
	}
}

type CommentRes struct {
	Replies    []*ReplyItem `json:"replies"`
	NextCursor uint64       `json:"next_cursor"`
	HasNext    bool         `json:"has_next"`
}

type GetSubCommentsReq struct {
	Oid    uint64 `form:"oid"`
	RootId uint64 `form:"root"`
	Cursor uint64 `form:"cursor,optional"`
}

func (r *GetSubCommentsReq) AsPb() *sdk.PageGetSubReplyReq {
	return &sdk.PageGetSubReplyReq{
		Oid:    r.Oid,
		RootId: r.RootId,
		Cursor: r.Cursor,
	}
}

type ReplyItem struct {
	*sdk.ReplyItem
	User *user.UserInfo `json:"user"`
}

type DetailedSubReply struct {
	Items      []*ReplyItem `json:"items"`
	NextCursor uint64       `json:"next_cursor"`
	HasNext    bool         `json:"has_next"`
}

// 带有子评论的评论信息
type DetailedReplyItem struct {
	Root       *ReplyItem        `json:"root"`
	SubReplies *DetailedSubReply `json:"sub_replies"`
}

func NewDetailedReplyItemFromPb(item *sdk.DetailedReplyItem, userMap map[string]*user.UserInfo) *DetailedReplyItem {
	details := &DetailedReplyItem{}
	details.Root = &ReplyItem{
		ReplyItem: item.Root,
	}
	if userMap != nil {
		details.Root.User = userMap[xconv.FormatUint(item.Root.Uid)]
	}

	details.SubReplies = &DetailedSubReply{
		Items:      make([]*ReplyItem, 0),
		HasNext:    item.Subreplies.HasNext,
		NextCursor: item.Subreplies.NextCursor,
	}
	for _, sub := range item.Subreplies.Items {
		item := &ReplyItem{
			ReplyItem: sub,
		}
		if userMap != nil {
			item.User = userMap[xconv.FormatUint(sub.Uid)]
		}

		details.SubReplies.Items = append(details.SubReplies.Items, item)
	}

	return details
}

type DetailedCommentRes struct {
	Replies    []*DetailedReplyItem `json:"replies"`
	PinReply   *DetailedReplyItem   `json:"pin_reply,omitempty"` // 置顶评论
	NextCursor uint64               `json:"next_cursor"`
	HasNext    bool                 `json:"has_next"`
}

// 删除评论
type DelReq struct {
	ReplyId uint64 `json:"reply_id"`
}

type PinAction int8

const (
	PinActionUnpin = 0
	PinActionPin   = 1
)

// 置顶评论
type PinReq struct {
	Oid     uint64    `json:"oid"`
	ReplyId uint64    `json:"reply_id"`
	Action  PinAction `json:"action"`
}

func (r *PinReq) Validate() error {
	if r.Action != PinActionUnpin && r.Action != PinActionPin {
		return errorx.ErrArgs.Msg("不支持的置顶操作")
	}

	return nil
}
