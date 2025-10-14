package note

import (
	"context"
	"net/http"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizfeed "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed"
	feedmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/feed/model"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	feedBiz   *bizfeed.Biz
	searchBiz *bizsearch.Biz
	userBiz   *bizuser.Biz
	notifyBiz *bizsysnotify.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		feedBiz:   bizz.FeedBiz,
		searchBiz: bizz.SearchBiz,
		userBiz:   bizz.UserBiz,
		notifyBiz: bizz.SysNotificationBiz,
	}
}

// 检查笔记是否存在
func (h *Handler) CheckNoteExistence(ctx context.Context, noteId int64) error {
	if resp, err := dep.NoteCreatorServer().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: noteId,
		}); err != nil {
		return err
	} else {
		if !resp.Exist {
			return errors.ErrNoteNotFound
		}
	}

	return nil
}

// 设置笔记的额外信息
func (h *Handler) assignNoteExtra(ctx context.Context, notes []*imodel.AdminNoteItem) {
	var (
		noteIds      = make([]int64, 0, len(notes))
		oidLiked     = make(map[int64]bool)
		oidCommented = make(map[int64]bool)
		uid          = metadata.Uid(ctx)
		eg           errgroup.Group
	)

	for _, n := range notes {
		noteIds = append(noteIds, int64(n.NoteId))
	}

	eg.Go(func() error {
		mappings := make(map[int64]*notev1.NoteIdList)
		mappings[uid] = &notev1.NoteIdList{
			NoteIds: noteIds,
		}

		// 点赞信息
		resp, err := dep.NoteInteractServer().BatchCheckUserLikeStatus(ctx,
			&notev1.BatchCheckUserLikeStatusRequest{
				Mappings: mappings,
			})
		if err != nil {
			return xerror.Wrapf(err, "failed to get user like status").WithCtx(ctx)
		}

		pairs := resp.GetResults()
		for _, likedInfo := range pairs[uid].GetList() {
			oidLiked[likedInfo.NoteId] = likedInfo.Liked
		}

		for _, note := range notes {
			noteId := int64(note.NoteId)
			note.Interact.Liked = oidLiked[noteId]
		}

		return nil
	})

	eg.Go(func() error {
		commentMappings := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
		commentMappings[uid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{
			Oids: noteIds,
		}
		// 评论信息
		resp, err := dep.Commenter().BatchCheckUserOnObject(ctx,
			&commentv1.BatchCheckUserOnObjectRequest{
				Mappings: commentMappings,
			})
		if err != nil {
			return xerror.Wrapf(err, "failed to get comment status").WithCtx(ctx)
		}

		// organize result
		pairs := resp.GetResults()
		for _, comInfo := range pairs[uid].GetList() {
			oidCommented[comInfo.Oid] = comInfo.Commented
		}
		for _, note := range notes {
			noteId := int64(note.NoteId)
			note.Interact.Commented = oidCommented[noteId]
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		xlog.Msgf("failed to assign note extra").Err(err).Errorx(ctx)
		return
	}

	for _, note := range notes {
		noteId := int64(note.NoteId)
		note.Interact.Liked = oidLiked[noteId]
		note.Interact.Commented = oidCommented[noteId]
	}
}

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid = metadata.Uid(r.Context())
		)

		req, err := xhttp.ParseValidate[LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		noteId := req.NoteId
		noteIdStr := noteId.String()
		ctx := r.Context()
		_, err = dep.NoteInteractServer().LikeNote(ctx, &notev1.LikeNoteRequest{
			NoteId:    int64(noteId),
			Uid:       uid,
			Operation: notev1.LikeNoteRequest_Operation(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "note.handler.likenote.synces",
			Job: func(ctx context.Context) error {
				var incr int64 = 1
				if req.Action == imodel.LikeReqActionUndo {
					incr = -1
				}

				err := h.searchBiz.NoteStatSyncer.AddLikeCount(ctx, noteIdStr, incr)
				if err != nil {
					xlog.Msg("note stat add like count failed").
						Extras("note_id", noteId, "note_id_str", noteIdStr).
						Err(err).Errorx(ctx)
				}

				return err
			},
		})

		xhttp.OkJson(w, nil)
	}
}

// 获取笔记点赞数量
func (h *Handler) GetNoteLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := int64(req.NoteId)
		resp, err := dep.NoteInteractServer().GetNoteLikes(r.Context(),
			&notev1.GetNoteLikesRequest{NoteId: nid})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetLikesRes{
			NoteId: imodel.NoteId(resp.NoteId),
			Count:  resp.Likes,
		})
	}
}

// 创建新标签
func (h *Handler) AddNewTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[AddTagReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := dep.NoteCreatorServer().AddTag(ctx,
			&notev1.AddTagRequest{
				Name: req.Name,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "add_new_tag",
			Job: func(newCtx context.Context) error {
				// 再查一遍
				newTag, err := dep.NoteFeedServer().GetTagInfo(newCtx,
					&notev1.GetTagInfoRequest{
						Id: resp.Id,
					})
				if err != nil {
					xlog.Msg("after adding new tag, get tag info failed").Extra("tag_id", resp.Id).Err(err).Errorx(newCtx)
					return err
				}

				tagId := imodel.TagId(newTag.Tag.Id).String()

				_, err = dep.DocumentServer().BatchAddNoteTag(newCtx, &searchv1.BatchAddNoteTagRequest{
					NoteTags: []*searchv1.NoteTag{
						{
							Id:    tagId,
							Name:  newTag.Tag.Name,
							Ctime: newTag.Tag.Ctime,
						},
					},
				})
				if err != nil {
					xlog.Msg("after adding new tag, failed to insert tag document").
						Extras("tag_id", resp.Id, "stag_id", tagId).
						Err(err).
						Errorx(newCtx)
					return err
				}

				return nil
			},
		})

		xhttp.OkJson(w, &AddTagRes{TagId: imodel.TagId(resp.Id)})
	}
}

// 搜索笔记标签
func (h *Handler) SearchTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SearchTagsReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := dep.SearchServer().SearchNoteTags(ctx, &searchv1.SearchNoteTagsRequest{
			Text: req.Name,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := make([]SearchedNoteTag, len(resp.Items))
		for idx, item := range resp.Items {
			if item == nil {
				continue
			}
			result[idx].Id = item.Id
			result[idx].Name = item.Name
		}

		xhttp.OkJson(w, result)
	}
}

// 获取用户点赞过的笔记
func (b *Handler) ListLikedNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetLikedNoteRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		if uid != req.Uid {
			// TODO 检查req.Uid的点赞记录是否公开
		}

		noteResp, err := dep.NoteInteractServer().PageListUserLikedNote(ctx,
			&notev1.PageListUserLikedNoteRequest{
				Uid:    req.Uid,
				Cursor: req.Cursor,
				Count:  req.Count,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp := &GetLikedNoteResponse{
			Items:      []*feedmodel.FeedNoteItem{},
			NextCursor: noteResp.NextCursor,
			HasNext:    noteResp.HasNext,
		}

		if len(noteResp.Items) > 0 {
			targets, err := b.feedBiz.AssembleNoteFeeds(ctx, noteResp.GetItems())
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}
			resp.Items = targets
		}

		xhttp.OkJson(w, resp)
	}
}
