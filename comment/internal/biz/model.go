package biz

import (
	"encoding/json"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
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

type ImageAssetMetadata struct {
	Width  uint32 `json:"w"`
	Height uint32 `json:"h"`
	Format string `json:"f"`
}

func makePbCommentImage(assets []*dao.CommentAsset) []*commentv1.CommentItemImage {
	imgs := make([]*commentv1.CommentItemImage, 0, len(assets))
	for _, img := range assets {
		var meta ImageAssetMetadata
		_ = json.Unmarshal(img.Metadata, &meta)
		imgs = append(imgs, &commentv1.CommentItemImage{
			Key: img.StoreKey,
			Meta: &commentv1.CommentItemImageMeta{
				Width:  meta.Width,
				Height: meta.Height,
				Format: meta.Format,
				Type:   model.CommentAssetType(img.Type).String(),
			},
		})
	}

	return imgs
}

func makeCommentAssetPO(commentId int64, images []*commentv1.CommentReqImage) []*dao.CommentAsset {
	assets := make([]*dao.CommentAsset, 0, len(images))
	for _, img := range images {
		meta := ImageAssetMetadata{
			Width:  img.Width,
			Height: img.Height,
			Format: img.Format,
		}
		metadata, _ := json.Marshal(&meta)
		assets = append(assets, &dao.CommentAsset{
			CommentId: commentId,
			Type:      model.CommentAssetImage,
			StoreKey:  img.StoreKey,
			Metadata:  metadata,
		})
	}

	return assets
}
