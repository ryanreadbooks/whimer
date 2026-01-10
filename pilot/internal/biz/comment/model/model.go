package model

import (
	"context"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type CommentImage struct {
	StoreKey string `json:"store_key"`
	Width    uint32 `json:"width"`
	Height   uint32 `json:"height"`
	Format   string `json:"format"`
}

// 与评论的交互信息
type CommentItemBaseInteract struct {
	Liked bool `json:"liked"` // 当前请求用户是否点赞了该评论
}

type CommentItemImageMetadata struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
	Type   string `json:"type"`
}

// 评论图片
type CommentItemImage struct {
	Url      string                   `json:"url"`
	Metadata CommentItemImageMetadata `json:"metadata"`
}

func NewCommentItemImageUrl(pbimg *commentv1.CommentItemImage) string {
	bucket := config.Conf.UploadResourceDefineMap["comment_image"].Bucket
	return imgproxy.GetSignedUrl(config.Conf.Oss.DisplayEndpointBucket(bucket),
		pbimg.GetKey(),
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality("50"),
	)
}

func NewCommentItemImageFromPb(pbimg *commentv1.CommentItemImage) *CommentItemImage {
	if pbimg == nil {
		return &CommentItemImage{}
	}

	return &CommentItemImage{
		Url: NewCommentItemImageUrl(pbimg),
		Metadata: CommentItemImageMetadata{
			Width:  pbimg.GetMeta().GetWidth(),
			Height: pbimg.GetMeta().GetHeight(),
			Format: pbimg.GetMeta().GetFormat(),
			Type:   pbimg.GetMeta().GetType(),
		},
	}
}

func NewCommentItemImageSliceFromPb(imgs []*commentv1.CommentItemImage) []*CommentItemImage {
	if imgs == nil {
		return []*CommentItemImage{}
	}

	items := make([]*CommentItemImage, 0, len(imgs))
	for _, img := range imgs {
		items = append(items, NewCommentItemImageFromPb(img))
	}

	return items
}

// 对客户端暴露的一条评论的结构
type CommentItemBase struct {
	Id        int64               `json:"id"`                 // 评论id
	Oid       model.NoteId        `json:"oid"`                // 被评论对象id
	Type      int32               `json:"type"`               // 评论类型
	Content   string              `json:"content"`            // 评论内容
	Uid       int64               `json:"uid"`                // 评论发表用户uid
	RootId    int64               `json:"root_id"`            // 根评论id
	ParentId  int64               `json:"parent_id"`          // 父评论id
	Ruid      int64               `json:"ruid"`               // 被回复的用户id
	LikeCount int64               `json:"like_count"`         // 点赞数
	HateCount int64               `json:"-"`                  // 点踩数
	Ctime     int64               `json:"ctime"`              // 发布时间
	Mtime     int64               `json:"mtime"`              // 修改时间
	Ip        string              `json:"-"`                  // ip
	IpLoc     string              `json:"ip_loc"`             // ip归属地
	IsPin     bool                `json:"is_pin"`             // 是否为置顶评论
	SubsCount int64               `json:"subs_count"`         // 子评论数
	Images    []*CommentItemImage `json:"images,omitempty"`   // 评论图片
	AtUsers   []*model.AtUser     `json:"at_users,omitempty"` // at_users 被@的用户列表

	Interact CommentItemBaseInteract `json:"interact"` // 交互信息
}

type CommentItem struct {
	*CommentItemBase
	User *userv1.UserInfo `json:"user"`
}

func NewCommentItemBaseFromPb(p *commentv1.CommentItem) *CommentItemBase {
	if p == nil {
		return &CommentItemBase{}
	}

	ctx := context.Background()
	ipLoc, _ := infra.Ip2Loc().Convert(ctx, p.Ip)

	atUsers := make([]*model.AtUser, 0, len(p.AtUsers))
	for _, a := range p.AtUsers {
		atUsers = append(atUsers, &model.AtUser{
			Nickname: a.Nickname,
			Uid:      a.Uid,
		})
	}

	return &CommentItemBase{
		Id:        p.Id,
		Oid:       model.NoteId(p.Oid),
		Type:      int32(p.Type),
		Content:   p.Content,
		Uid:       p.Uid,
		RootId:    p.RootId,
		ParentId:  p.ParentId,
		Ruid:      p.Uid,
		LikeCount: p.LikeCount,
		HateCount: p.HateCount,
		Mtime:     p.Mtime,
		Ctime:     p.Ctime,
		Ip:        p.Ip,
		IpLoc:     ipLoc,
		IsPin:     p.IsPin,
		SubsCount: p.SubsCount,
		Images:    NewCommentItemImageSliceFromPb(p.Images),
		AtUsers:   atUsers,
	}
}

type DetailedSubComment struct {
	Items      []*CommentItem `json:"items"`
	NextCursor int64          `json:"next_cursor"`
	HasNext    bool           `json:"has_next"`
}

// 带有子评论的评论信息
type DetailedCommentItem struct {
	Root        *CommentItem        `json:"root"`
	SubComments *DetailedSubComment `json:"sub_comments"`
}

func NewDetailedCommentItemFromPb(item *commentv1.DetailedCommentItem,
	userMap map[string]*userv1.UserInfo) *DetailedCommentItem {

	details := &DetailedCommentItem{}
	details.Root = &CommentItem{
		CommentItemBase: NewCommentItemBaseFromPb(item.Root),
	}
	if userMap != nil {
		details.Root.User = userMap[xconv.FormatInt(item.Root.Uid)]
	}

	details.SubComments = &DetailedSubComment{
		Items:      make([]*CommentItem, 0),
		HasNext:    item.SubComments.HasNext,
		NextCursor: item.SubComments.NextCursor,
	}
	for _, sub := range item.SubComments.Items {
		item := &CommentItem{
			CommentItemBase: NewCommentItemBaseFromPb(sub),
		}
		if userMap != nil {
			item.User = userMap[xconv.FormatInt(sub.Uid)]
		}

		details.SubComments.Items = append(details.SubComments.Items, item)
	}

	return details
}
