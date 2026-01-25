package convert

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/vo"
)

func PbCommentItemToEntity(p *commentv1.CommentItem) *entity.Comment {
	if p == nil {
		return nil
	}

	images := make([]*entity.CommentImage, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, &entity.CommentImage{
			Key:    img.GetKey(),
			Width:  img.GetMeta().GetWidth(),
			Height: img.GetMeta().GetHeight(),
			Format: img.GetMeta().GetFormat(),
			Type:   img.GetMeta().GetType(),
		})
	}

	atUsers := make([]*entity.AtUser, 0, len(p.AtUsers))
	for _, au := range p.AtUsers {
		atUsers = append(atUsers, &entity.AtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	return &entity.Comment{
		Id:        p.Id,
		Oid:       p.Oid,
		Type:      int32(p.Type),
		Content:   p.Content,
		Uid:       p.Uid,
		RootId:    p.RootId,
		ParentId:  p.ParentId,
		Ruid:      p.Ruid,
		LikeCount: p.LikeCount,
		HateCount: p.HateCount,
		Ctime:     p.Ctime,
		Mtime:     p.Mtime,
		Ip:        p.Ip,
		IsPin:     p.IsPin,
		SubsCount: p.SubsCount,
		Images:    images,
		AtUsers:   atUsers,
	}
}

func PbDetailedCommentItemToEntity(p *commentv1.DetailedCommentItem) *entity.DetailedComment {
	if p == nil {
		return nil
	}

	subItems := make([]*entity.Comment, 0, len(p.SubComments.GetItems()))
	for _, item := range p.SubComments.GetItems() {
		subItems = append(subItems, PbCommentItemToEntity(item))
	}

	return &entity.DetailedComment{
		Root: PbCommentItemToEntity(p.Root),
		SubComments: &entity.SubComments{
			Items:      subItems,
			NextCursor: p.SubComments.GetNextCursor(),
			HasNext:    p.SubComments.GetHasNext(),
		},
	}
}

func PinActionAsPb(action vo.PinAction) commentv1.CommentAction {
	switch action {
	case vo.PinActionPin:
		return commentv1.CommentAction_COMMENT_ACTION_DO
	default:
		return commentv1.CommentAction_COMMENT_ACTION_UNDO
	}
}

func ThumbActionAsPb(action vo.ThumbAction) commentv1.CommentAction {
	switch action {
	case vo.ThumbActionDo:
		return commentv1.CommentAction_COMMENT_ACTION_DO
	default:
		return commentv1.CommentAction_COMMENT_ACTION_UNDO
	}
}
