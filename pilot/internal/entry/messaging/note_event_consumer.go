package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xretry"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"

	noteevent "github.com/ryanreadbooks/whimer/note/pkg/event/note"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

func startNoteEventConsumer(bizz *biz.Biz) {
	backoff := xretry.NewPermenantBackoff(time.Millisecond*100, time.Second*8, 2.0)
	concurrent.SafeGo2(rootCtx, concurrent.SafeGo2Opt{
		Name: "pilot.note_event.consumer",
		Job: func(ctx context.Context) error {
			xlog.Msg("start consuming note event")

			trapped := false
			for {
				kMsg, err := noteEventConsumer.ReadMessage(ctx)
				if err != nil {
					// 区分不同类型的错误
					if errors.Is(err, context.Canceled) {
						// 上下文取消才认为是退出
						xlog.Msg("pilot note.event.consumer context done").Err(err).Info()
						break
					}

					xlog.Msg("pilot note.event.consumer read message err, will retry").Err(err).Errorx(ctx)

					rest, _ := backoff.NextBackOff()
					trapped = true
					select {
					case <-time.After(rest):
					case <-ctx.Done():
						// check again
						xlog.Msg("pilot note.event.consumer context done during retry delay").Info()
						break
					}
					continue
				}

				if trapped {
					backoff.Success()
					trapped = false
				}

				var ev noteevent.NoteEvent
				err = json.Unmarshal(kMsg.Value, &ev)
				if err != nil {
					xlog.Msg("pilot note.event.consumer json.Unmarshal err").Err(err).Errorx(ctx)
					continue
				}

				msgCtx := xkafka.ContextFromKafkaHeaders(kMsg.Headers)
				if err := handleNoteEvent(msgCtx, bizz, ev); err != nil {
					xlog.Msg("pilot note.event.consumer handle note event err").Err(err).Errorx(ctx)
				}
			}

			return nil
		},
	})
}

func handleNoteEvent(ctx context.Context, bizz *biz.Biz, ev noteevent.NoteEvent) error {
	switch ev.Type {
	case noteevent.NotePublished:
		// 解码payload
		var data noteevent.NotePublishedEventData
		if err := json.Unmarshal(ev.Payload, &data); err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note published event failed to unmarshal payload").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}

		owner, err := bizz.UserBiz.GetUser(ctx, data.Note.Owner)
		if err != nil {
			xlog.Msg("handle note event get username failed").Extra("user_id", data.Note.Owner).Errorx(ctx)
		}

		var assetType searchv1.Note_AssetType
		switch data.Note.Type {
		case "image":
			assetType = searchv1.Note_ASSET_TYPE_IMAGE
		case "video":
			assetType = searchv1.Note_ASSET_TYPE_VIDEO
		}

		tagList := []*searchv1.NoteTag{}
		for _, t := range data.Note.Tags {
			tagList = append(tagList, &searchv1.NoteTag{
				Id:    t.Tid,
				Name:  t.Name,
				Ctime: t.Ctime,
			})
		}

		docNote := searchv1.Note{
			NoteId:   data.Note.Nid,
			Title:    data.Note.Title,
			Desc:     data.Note.Desc,
			CreateAt: data.Note.Ctime,
			UpdateAt: data.Note.Utime,
			Author: &searchv1.Note_Author{
				Uid:      data.Note.Owner,
				Nickname: owner.Nickname,
			},
			TagList:    tagList,
			AssetType:  assetType,
			Visibility: searchv1.Note_VISIBILITY_PUBLIC,
		}

		return bizz.SearchBiz.AddNoteDoc(ctx, &docNote)
	case noteevent.NoteDeleted:
		var data noteevent.NoteDeletedEventData
		if err := json.Unmarshal(ev.Payload, &data); err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note deleted event failed to unmarshal payload").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}

		return bizz.SearchBiz.DeleteNoteDoc(ctx, data.Note.Nid)
	default:
		// TODO 对其它类型不处理

	}

	return nil
}
