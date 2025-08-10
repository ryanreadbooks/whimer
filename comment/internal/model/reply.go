package model

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/xnet"
)

// 评论类型
type ReplyType int8

const (
	ReplyText      ReplyType = 0
	ReplyImageText ReplyType = 1
)

// 评论是否置顶
const (
	IsNotPinned = 0
	IsPinned    = 1
)

// 评论状态
type ReplyState int8

// 评论状态
const (
	// TODO define more reply state
	ReplyStateNormal ReplyState = 0
)

const (
	minContentLen = 1
	maxContentLen = 2000
)

type ReplyItem struct {
	Id         int64  `json:"id"`
	Oid        int64  `json:"oid"`
	ReplyType  int8   `json:"reply_type"`
	Content    string `json:"content"`
	Uid        int64  `json:"uid"`
	RootId     int64  `json:"root_id"`
	ParentId   int64  `json:"parent_id"`
	RepliedUid int64  `json:"replied_uid"`
	Ctime      int64  `json:"ctime"`
	Mtime      int64  `json:"mtime"`
	Ip         string `json:"ip"`
	IsPin      bool   `json:"is_pin"`

	// 下面的字段需要额外填充
	LikeCount int64 `json:"like_count"`
	HateCount int64 `json:"hate_count"`
	SubsCount int64 `json:"subs_count"` // 其下子评论的数量
}

func NewReplyItem(d *dao.Comment) *ReplyItem {
	return &ReplyItem{
		Id:         d.Id,
		Oid:        d.Oid,
		ReplyType:  d.CType,
		Content:    d.Content,
		Uid:        d.Uid,
		RootId:     d.RootId,
		ParentId:   d.ParentId,
		RepliedUid: d.ReplyUid,
		LikeCount:  int64(d.Like),
		HateCount:  int64(d.Dislike),
		Ctime:      d.Ctime,
		Mtime:      d.Mtime,
		Ip:         xnet.IntAsIp(uint32(d.Ip)),
		IsPin:      d.IsPin == dao.AlreadyPinned,
	}
}

func (r *ReplyItem) IsRoot() bool {
	return r.RootId == 0 && r.ParentId == 0
}

func (r *ReplyItem) AsPb() *commentv1.ReplyItem {
	return &commentv1.ReplyItem{
		Id:        r.Id,
		Oid:       r.Oid,
		ReplyType: uint32(r.ReplyType),
		Content:   r.Content,
		Uid:       r.Uid,
		RootId:    r.RootId,
		ParentId:  r.ParentId,
		Ruid:      r.RepliedUid,
		LikeCount: r.LikeCount,
		HateCount: r.HateCount,
		Ctime:     r.Ctime,
		Mtime:     r.Mtime,
		Ip:        r.Ip,
		IsPin:     r.IsPin,
		SubsCount: r.SubsCount,
	}
}

type DetailedReplyItem struct {
	Root *ReplyItem   // 主评论本身
	Subs *PageReplies // 主评论其下子评论
}

func (r *DetailedReplyItem) AsPb() *commentv1.DetailedReplyItem {
	return &commentv1.DetailedReplyItem{
		Root: r.Root.AsPb(),
		SubReplies: &commentv1.DetailedSubReply{
			Items:      ItemsAsPbs(r.Subs.Items),
			NextCursor: r.Subs.NextCursor,
			HasNext:    r.Subs.HasNext,
		},
	}
}

type PageReplies struct {
	Items      []*ReplyItem
	NextCursor int64
	HasNext    bool
}

type PageRepliesWithTotal struct {
	Items []*ReplyItem
	Total int64
}

type DetailedReplyItemV2 struct {
	Root *ReplyItem            // 主评论本身
	Subs *PageRepliesWithTotal // 主评论其下子评论
}

type PageDetailedRepliesV2 struct {
	Items      []*DetailedReplyItemV2
	NextCursor int64
	HasNext    bool
}

type PageDetailedReplies struct {
	Items      []*DetailedReplyItem
	NextCursor int64
	HasNext    bool
}

func ItemsAsPbs(rs []*ReplyItem) []*commentv1.ReplyItem {
	r := make([]*commentv1.ReplyItem, 0, len(rs))
	for _, item := range rs {
		r = append(r, item.AsPb())
	}
	return r
}

func DetailedItemsAsPbs(rs []*DetailedReplyItem) []*commentv1.DetailedReplyItem {
	r := make([]*commentv1.DetailedReplyItem, 0, len(rs))
	for _, item := range rs {
		r = append(r, &commentv1.DetailedReplyItem{
			Root: item.Root.AsPb(),
			SubReplies: &commentv1.DetailedSubReply{
				Items:      ItemsAsPbs(item.Subs.Items),
				NextCursor: item.Subs.NextCursor,
				HasNext:    item.Subs.HasNext,
			},
		})
	}
	return r
}

func DetailedItemsV2AsPbs(rs []*DetailedReplyItemV2) []*commentv1.DetailedReplyItemV2 {
	r := make([]*commentv1.DetailedReplyItemV2, 0, len(rs))
	for _, item := range rs {
		r = append(r, &commentv1.DetailedReplyItemV2{
			Root: item.Root.AsPb(),
			SubReplies: &commentv1.DetailedSubReplyV2{
				Items: ItemsAsPbs(item.Subs.Items),
				Total: item.Subs.Total,
			},
		})
	}
	return r
}

func IsRoot(rootId, parentId int64) bool {
	return rootId == 0 && parentId == 0
}

type UidCommentOnOid struct {
	Uid       int64
	Oid       int64
	Commented bool
}

func (o *UidCommentOnOid) AsPb() *commentv1.OidCommented {
	return &commentv1.OidCommented{
		Oid:       o.Oid,
		Commented: o.Commented,
	}
}
