package comment

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
)

type PubReq struct {
	ReplyType uint32 `json:"reply_type"`
	Oid       uint64 `json:"oid"`
	Content   string `json:"content"`
	RootId    uint64 `json:"root_id,omitempty,optional"`
	ParentId  uint64 `json:"parent_id,omitempty,optional"`
	ReplyUid  uint64 `json:"reply_uid"`
}

func (r *PubReq) AsPb() *commentv1.AddReplyRequest {
	return &commentv1.AddReplyRequest{
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

func (r *GetCommentsReq) AsPb() *commentv1.PageGetReplyRequest {
	return &commentv1.PageGetReplyRequest{
		Oid:    r.Oid,
		Cursor: r.Cursor,
		SortBy: commentv1.SortType(r.SortBy),
	}
}

func (r *GetCommentsReq) AsDetailedPb() *commentv1.PageGetDetailedReplyRequest {
	return &commentv1.PageGetDetailedReplyRequest{
		Oid:    r.Oid,
		Cursor: r.Cursor,
		SortBy: commentv1.SortType(r.SortBy),
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

func (r *GetSubCommentsReq) AsPb() *commentv1.PageGetSubReplyRequest {
	return &commentv1.PageGetSubReplyRequest{
		Oid:    r.Oid,
		RootId: r.RootId,
		Cursor: r.Cursor,
	}
}

type ReplyItem struct {
	*commentv1.ReplyItem
	User *userv1.UserInfo `json:"user"`
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

func NewDetailedReplyItemFromPb(item *commentv1.DetailedReplyItem, userMap map[string]*userv1.UserInfo) *DetailedReplyItem {
	details := &DetailedReplyItem{}
	details.Root = &ReplyItem{
		ReplyItem: item.Root,
	}
	if userMap != nil {
		details.Root.User = userMap[xconv.FormatUint(item.Root.Uid)]
	}

	details.SubReplies = &DetailedSubReply{
		Items:      make([]*ReplyItem, 0),
		HasNext:    item.SubReplies.HasNext,
		NextCursor: item.SubReplies.NextCursor,
	}
	for _, sub := range item.SubReplies.Items {
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
		return xerror.ErrArgs.Msg("不支持的置顶操作")
	}

	return nil
}

type ThumbAction uint8

const (
	ThumbActionUndo ThumbAction = ThumbAction(commentv1.ReplyAction_REPLY_ACTION_UNDO) // 取消
	ThumbActionDo   ThumbAction = ThumbAction(commentv1.ReplyAction_REPLY_ACTION_DO)   // 执行
)

type thumbActionChecker struct{}

func (c thumbActionChecker) check(action ThumbAction) error {
	if action != ThumbActionUndo && action != ThumbActionDo {
		return xerror.ErrArgs.Msg("不支持的操作")
	}

	return nil
}

// 点赞评论/取消点赞评论
type ThumbUpReq struct {
	thumbActionChecker
	ReplyId uint64      `json:"reply_id"`
	Action  ThumbAction `json:"action"`
}

func (r *ThumbUpReq) Validate() error {
	return r.check(r.Action)
}

// 点踩评论/取消点踩评论
type ThumbDownReq struct {
	thumbActionChecker
	ReplyId uint64      `json:"reply_id"`
	Action  ThumbAction `json:"action"`
}

func (r *ThumbDownReq) Validate() error {
	return r.check(r.Action)
}

type GetLikeCountReq struct {
	ReplyId uint64 `form:"reply_id"`
}

func (r *GetLikeCountReq) Validate() error {
	if r.ReplyId <= 0 {
		return xerror.ErrArgs.Msg("评论不存在")
	}

	return nil
}

type GetLikeCountRes struct {
	ReplyId uint64 `json:"rid"`
	Likes   uint64 `json:"likes"`
}
