package model

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/xnet"
)

// 一条评论的数据模型
type CommentItem struct {
	Id         int64                         `json:"id"`
	Oid        int64                         `json:"oid"`
	Type       int8                          `json:"type"`
	Content    string                        `json:"content"`
	Uid        int64                         `json:"uid"`
	RootId     int64                         `json:"root_id"`
	ParentId   int64                         `json:"parent_id"`
	RepliedUid int64                         `json:"replied_uid"`
	Ctime      int64                         `json:"ctime"`
	Mtime      int64                         `json:"mtime"`
	Ip         string                        `json:"ip"`
	IsPin      bool                          `json:"is_pin"`
	Images     []*commentv1.CommentItemImage `json:"images"`

	// 下面的字段需要额外填充
	LikeCount int64 `json:"like_count"`
	HateCount int64 `json:"hate_count"`
	SubsCount int64 `json:"subs_count"` // 其下子评论的数量
}

func NewCommentItemFromDao(d *dao.Comment) *CommentItem {
	return &CommentItem{
		Id:         d.Id,
		Oid:        d.Oid,
		Type:       d.Type,
		Content:    d.Content,
		Uid:        d.Uid,
		RootId:     d.RootId,
		ParentId:   d.ParentId,
		RepliedUid: d.ReplyUid,
		LikeCount:  int64(d.Like),
		HateCount:  int64(d.Dislike),
		Ctime:      d.Ctime,
		Mtime:      d.Mtime,
		Ip:         xnet.BytesIpAsString(d.Ip),
		IsPin:      d.IsPin == dao.AlreadyPinned,
	}
}

func (r *CommentItem) IsRoot() bool {
	return r.RootId == 0 && r.ParentId == 0
}

func (r *CommentItem) AsPb() *commentv1.CommentItem {
	return &commentv1.CommentItem{
		Id:        r.Id,
		Oid:       r.Oid,
		Type:      commentv1.CommentType(r.Type),
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
		Images:    r.Images,
	}
}

type DetailedCommentItem struct {
	Root *CommentItem  // 主评论本身
	Subs *PageComments // 主评论其下子评论
}

func (r *DetailedCommentItem) AsPb() *commentv1.DetailedCommentItem {
	return &commentv1.DetailedCommentItem{
		Root: r.Root.AsPb(),
		SubComments: &commentv1.DetailedSubComment{
			Items:      ItemsAsPbs(r.Subs.Items),
			NextCursor: r.Subs.NextCursor,
			HasNext:    r.Subs.HasNext,
		},
	}
}

type PageComments struct {
	Items      []*CommentItem
	NextCursor int64
	HasNext    bool
}

type PageCommentsWithTotal struct {
	Items []*CommentItem
	Total int64
}

type DetailedCommentItemV2 struct {
	Root *CommentItem           // 主评论本身
	Subs *PageCommentsWithTotal // 主评论其下子评论
}

type PageDetailedCommentsV2 struct {
	Items      []*DetailedCommentItemV2
	NextCursor int64
	HasNext    bool
}

type PageDetailedComments struct {
	Items      []*DetailedCommentItem
	NextCursor int64
	HasNext    bool
}

func ItemsAsPbs(rs []*CommentItem) []*commentv1.CommentItem {
	r := make([]*commentv1.CommentItem, 0, len(rs))
	for _, item := range rs {
		r = append(r, item.AsPb())
	}
	return r
}

func DetailedItemsAsPbs(rs []*DetailedCommentItem) []*commentv1.DetailedCommentItem {
	r := make([]*commentv1.DetailedCommentItem, 0, len(rs))
	for _, item := range rs {
		r = append(r, &commentv1.DetailedCommentItem{
			Root: item.Root.AsPb(),
			SubComments: &commentv1.DetailedSubComment{
				Items:      ItemsAsPbs(item.Subs.Items),
				NextCursor: item.Subs.NextCursor,
				HasNext:    item.Subs.HasNext,
			},
		})
	}
	return r
}

func DetailedItemsV2AsPbs(rs []*DetailedCommentItemV2) []*commentv1.DetailedCommentItemV2 {
	r := make([]*commentv1.DetailedCommentItemV2, 0, len(rs))
	for _, item := range rs {
		r = append(r, &commentv1.DetailedCommentItemV2{
			Root: item.Root.AsPb(),
			SubComments: &commentv1.DetailedSubCommentV2{
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
