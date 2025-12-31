package note

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/note/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

// AddTag 创建新标签
func (b *Biz) AddTag(ctx context.Context, name string) (*model.AddTagRes, error) {
	resp, err := dep.NoteCreatorServer().AddTag(ctx, &notev1.AddTagRequest{Name: name})
	if err != nil {
		return nil, err
	}

	return &model.AddTagRes{TagId: imodel.TagId(resp.Id)}, nil
}

// AsyncTagToSearcher 同步标签到搜索引擎
func (b *Biz) AsyncTagToSearcher(ctx context.Context, tagId int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "sync_tag_to_searcher",
		Job: func(ctx context.Context) error {
			newTag, err := dep.NoteFeedServer().GetTagInfo(ctx, &notev1.GetTagInfoRequest{Id: tagId})
			if err != nil {
				xlog.Msg("after adding new tag, get tag info failed").Extra("tag_id", tagId).Err(err).Errorx(ctx)
				return err
			}

			tagIdStr := imodel.TagId(newTag.Tag.Id).String()
			_, err = dep.DocumentServer().BatchAddNoteTag(ctx, &searchv1.BatchAddNoteTagRequest{
				NoteTags: []*searchv1.NoteTag{{
					Id:    tagIdStr,
					Name:  newTag.Tag.Name,
					Ctime: newTag.Tag.Ctime,
				}},
			})
			if err != nil {
				xlog.Msg("after adding new tag, failed to insert tag document").
					Extras("tag_id", tagId, "stag_id", tagIdStr).Err(err).Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

// SearchTags 搜索标签
func (b *Biz) SearchTags(ctx context.Context, name string) ([]model.SearchedTag, error) {
	resp, err := dep.SearchServer().SearchNoteTags(ctx, &searchv1.SearchNoteTagsRequest{Text: name})
	if err != nil {
		return nil, err
	}

	result := make([]model.SearchedTag, len(resp.Items))
	for idx, item := range resp.Items {
		if item == nil {
			continue
		}
		result[idx].Id = item.Id
		result[idx].Name = item.Name
	}

	return result, nil
}
