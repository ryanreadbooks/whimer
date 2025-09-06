package feed

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 搜索笔记
func (h *Handler) SearchNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[SearchNotesReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		filters := make([]*searchv1.NoteFilter, 0)
		for _, f := range req.Filters {
			filters = append(filters, &searchv1.NoteFilter{
				Type:  searchv1.NoteFilterType(searchv1.NoteFilterType_value[f.Type]),
				Value: f.Value,
			})
		}

		searchingResp, err := infra.SearchServer().SearchNotes(ctx,
			&searchv1.SearchNotesRequest{
				Keyword:   req.Keyword,
				PageToken: req.PageToken,
				Count:     req.Count,
				Filters:   filters,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		noteIds := searchingResp.GetNoteIds() // string类型的 mixed的
		nids := []int64{}
		for _, n := range noteIds {
			var id model.NoteId
			err = id.UnmarshalText([]byte(n))
			if err == nil {
				nids = append(nids, int64(id))
			} else {
				xlog.Msgf("search notes failed to parse note id").Extra("note_id", n).Err(err).Errorx(ctx)
			}
		}

		resp := &SearchNotesRes{}
		if len(noteIds) != 0 {
			xlog.Msgf("search notes got %v", noteIds).Debugx(ctx)
			notes, err := h.bizz.BatchGetNote(ctx, nids)
			if err != nil {
				xhttp.Error(r, w, err)
				return
			}

			resp.HasNext = searchingResp.HasNext
			resp.Total = searchingResp.Total
			resp.NextToken = searchingResp.NextToken
			resp.Items = notes
		}

		xhttp.OkJson(w, resp)
	}
}

// 获取搜索可用的过滤器
func (h *Handler) GetSearchNotesAvailableFilters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
