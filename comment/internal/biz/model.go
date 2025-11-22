package biz

import (
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xnet"
)

func NewCommentItemFromDao(d *dao.Comment) *model.CommentItem {
	return &model.CommentItem{
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

func NewCommentItemSliceFromDao(ds []*dao.Comment) []*model.CommentItem {
	result := make([]*model.CommentItem, 0, len(ds))
	for _, d := range ds {
		result = append(result, NewCommentItemFromDao(d))
	}

	return result
}
