package svc

import (
	"context"
	"errors"

	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/oss"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	noterepo "github.com/ryanreadbooks/whimer/note/internal/infra/repo/note"
	notemodel "github.com/ryanreadbooks/whimer/note/internal/model/note"
)

// 一些通用的组件函数

// 判断笔记是否存在
func IsNoteExist(ctx context.Context, nid uint64) (bool, error) {
	if nid <= 0 {
		return false, nil
	}

	_, err := infra.Repo().NoteRepo.FindOne(ctx, nid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return false, xerror.Wrapf(err, "note repo find one failed").WithExtra("noteId", nid).WithCtx(ctx)
		}
		return false, nil
	}

	return true, nil
}

// 组装note信息
func AssembleNotes(ctx context.Context, notes []*noterepo.Model) (*notemodel.BatchNoteItem, error) {
	var noteIds = make([]uint64, 0, len(notes))
	likesReq := make([]*counterv1.GetSummaryRequest, 0, len(notes))
	for _, note := range notes {
		noteIds = append(noteIds, note.Id)
		likesReq = append(likesReq, &counterv1.GetSummaryRequest{
			BizCode: global.NoteLikeBizcode,
			Oid:     note.Id,
		})
	}

	// 获取资源信息
	noteAssets, err := infra.Repo().NoteAssetRepo.FindByNoteIds(ctx, noteIds)
	if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
		return nil, xerror.Wrapf(err, "repo note asset failed")
	}

	// 组合notes和noteAssets
	var res notemodel.BatchNoteItem
	for _, note := range notes {
		item := &notemodel.Item{
			NoteId:   note.Id,
			Title:    note.Title,
			Desc:     note.Desc,
			Privacy:  note.Privacy,
			CreateAt: note.CreateAt,
			UpdateAt: note.UpdateAt,
		}
		for _, asset := range noteAssets {
			if note.Id == asset.NoteId {
				item.Images = append(item.Images, &notemodel.ItemImage{
					Url: oss.GetPublicVisitUrl(
						config.Conf.Oss.Bucket,
						asset.AssetKey,
						config.Conf.Oss.DisplayEndpoint,
					),
					Type: int(asset.AssetType),
				})
			}
		}

		res.Items = append(res.Items, item)
	}

	// 获取点赞数量
	AssignNoteLikes(ctx, &res)

	return &res, nil
}

// 获取笔记点赞信息
func AssignNoteLikes(ctx context.Context, batch *notemodel.BatchNoteItem) (*notemodel.BatchNoteItem, error) {
	var (
		notes   = batch.Items
		noteIds = make([]uint64, 0, len(notes))
		reqs    = make([]*counterv1.GetSummaryRequest, 0, len(notes))
	)

	for _, note := range notes {
		noteIds = append(noteIds, note.NoteId)
		reqs = append(reqs, &counterv1.GetSummaryRequest{
			BizCode: global.NoteLikeBizcode,
			Oid:     note.NoteId,
		})
	}

	// 获取点赞数量
	resp, err := dep.GetCounter().BatchGetSummary(ctx, &counterv1.BatchGetSummaryRequest{
		Requests: reqs,
	})
	if err != nil {
		// 仅打印日志不返回error
		xlog.Msg("counter failed to batch get summary").
			Err(err).
			Extra("note_ids", noteIds).
			Infox(ctx)
	}

	if resp != nil {
		m := make(map[uint64]uint64, len(resp.Responses))
		for _, r := range resp.Responses {
			m[r.Oid] = r.Count
		}
		// 赋值
		for _, item := range batch.Items {
			if likeCnt, ok := m[item.NoteId]; ok {
				item.Likes = likeCnt
			}
		}
	}

	return batch, nil
}

// 数据库中获取笔记
func GetNote(ctx context.Context, noteId uint64) (*noterepo.Model, error) {
	note, err := CacheGetNote(ctx, noteId)
	if err != nil {
		note, err = infra.Repo().NoteRepo.FindOne(ctx, noteId)
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrNoteNotFound
		}
		if err != nil {
			return nil, xerror.Wrapf(err, "repo note find one failed").WithCtx(ctx)
		}

		concurrent.SafeGo(func() {
			ctxc := context.WithoutCancel(ctx)
			if errg := CacheSetNote(ctxc, note); errg != nil {
				xlog.Msg("cache set note failed").Err(err).Extra("note", note).Errorx(ctxc)
			}
		})
	}

	return note, nil
}

// 获取用户是否点赞过笔记
func CheckUserLikeStatus(ctx context.Context, uid, noteId uint64) (bool, error) {
	resp, err := dep.GetCounter().GetRecord(ctx, &counterv1.GetRecordRequest{
		BizCode: global.NoteLikeBizcode,
		Uid:     uid,
		Oid:     noteId,
	})
	if err != nil {
		return false, xerror.Wrapf(err, "CheckUserLikeStatus counter get record failed").
			WithExtra("noteId", noteId).
			WithExtra("user", uid).
			WithCtx(ctx)
	}

	return resp.GetRecord().GetAct() == counterv1.RecordAct_RECORD_ACT_ADD, nil
}
