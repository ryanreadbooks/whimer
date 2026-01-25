package dto

import (
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/entity"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

// 评论图片
type CommentImage struct {
	Url      string               `json:"url"`
	Metadata CommentImageMetadata `json:"metadata"`
}

type CommentImageMetadata struct {
	Width  uint32 `json:"width"`
	Height uint32 `json:"height"`
	Format string `json:"format"`
	Type   string `json:"type"`
}

// 与评论的交互信息
type CommentInteract struct {
	Liked bool `json:"liked"`
}

func UserVoToCommentUser(u *uservo.User) *commondto.User {
	if u == nil {
		return nil
	}
	return &commondto.User{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		StyleSign: u.StyleSign,
	}
}

// 评论项
type Comment struct {
	Id        int64               `json:"id"`
	Oid       int64               `json:"oid"`
	Type      int32               `json:"type"`
	Content   string              `json:"content"`
	Uid       int64               `json:"uid"`
	RootId    int64               `json:"root_id"`
	ParentId  int64               `json:"parent_id"`
	Ruid      int64               `json:"ruid"`
	LikeCount int64               `json:"like_count"`
	Ctime     int64               `json:"ctime"`
	Mtime     int64               `json:"mtime"`
	Ip        string              `json:"-"`      // 原始 IP，不对外暴露
	IpLoc     string              `json:"ip_loc"` // IP 归属地
	IsPin     bool                `json:"is_pin"`
	SubsCount int64               `json:"subs_count"`
	Images    []*CommentImage     `json:"images,omitempty"`
	AtUsers   []*commondto.AtUser `json:"at_users,omitempty"`

	User     *commondto.User `json:"user"`
	Interact CommentInteract `json:"interact"`
}

// 子评论列表
type SubComments struct {
	Items      []*Comment `json:"items"`
	NextCursor int64      `json:"next_cursor"`
	HasNext    bool       `json:"has_next"`
}

// 带子评论的评论
type DetailedComment struct {
	Root        *Comment     `json:"root"`
	SubComments *SubComments `json:"sub_comments"`
}

// 评论列表响应
type CommentListResult struct {
	Items      []*Comment `json:"items"`
	NextCursor int64      `json:"next_cursor"`
	HasNext    bool       `json:"has_next"`
}

// 带子评论的评论列表响应
type DetailedCommentListResult struct {
	Comments   []*DetailedComment `json:"comments"`
	PinComment *DetailedComment   `json:"pin_comment,omitempty"`
	NextCursor int64              `json:"next_cursor"`
	HasNext    bool               `json:"has_next"`
}

// 点赞数响应
type LikeCountResult struct {
	CommentId int64 `json:"comment_id"`
	Likes     int64 `json:"likes"`
}

// 评论内容
type CommentContent struct {
	Text    string
	AtUsers []commondto.AtUser
}

// 转换函数
func EntityToComment(e *entity.Comment) *Comment {
	if e == nil {
		return nil
	}

	images := make([]*CommentImage, 0, len(e.Images))
	for _, img := range e.Images {
		images = append(images, entityImageToDto(img))
	}

	atUsers := make([]*commondto.AtUser, 0, len(e.AtUsers))
	for _, au := range e.AtUsers {
		atUsers = append(atUsers, &commondto.AtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	return &Comment{
		Id:        e.Id,
		Oid:       e.Oid,
		Type:      e.Type,
		Content:   e.Content,
		Uid:       e.Uid,
		RootId:    e.RootId,
		ParentId:  e.ParentId,
		Ruid:      e.Ruid,
		LikeCount: e.LikeCount,
		Ctime:     e.Ctime,
		Mtime:     e.Mtime,
		Ip:        e.Ip, // 保留原始 IP，由 service 层转换
		IsPin:     e.IsPin,
		SubsCount: e.SubsCount,
		Images:    images,
		AtUsers:   atUsers,
	}
}

func entityImageToDto(e *entity.CommentImage) *CommentImage {
	if e == nil {
		return nil
	}

	bucket := config.Conf.UploadResourceDefineMap["comment_image"].Bucket
	url := imgproxy.GetSignedUrl(config.Conf.Oss.DisplayEndpointBucket(bucket),
		e.Key,
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
		imgproxy.WithQuality("50"),
	)

	return &CommentImage{
		Url: url,
		Metadata: CommentImageMetadata{
			Width:  e.Width,
			Height: e.Height,
			Format: e.Format,
			Type:   e.Type,
		},
	}
}

func EntityToDetailedComment(e *entity.DetailedComment) *DetailedComment {
	if e == nil {
		return nil
	}

	subItems := make([]*Comment, 0, len(e.SubComments.Items))
	for _, item := range e.SubComments.Items {
		subItems = append(subItems, EntityToComment(item))
	}

	return &DetailedComment{
		Root: EntityToComment(e.Root),
		SubComments: &SubComments{
			Items:      subItems,
			NextCursor: e.SubComments.NextCursor,
			HasNext:    e.SubComments.HasNext,
		},
	}
}
