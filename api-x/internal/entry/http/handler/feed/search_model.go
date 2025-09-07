package feed

import "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed/model"

type SearchNotesFilter struct {
	Type  string `json:"type,optional"`
	Value string `json:"value,optional"`
}

type SearchNotesReq struct {
	Keyword   string              `json:"keyword"`
	PageToken string              `json:"page_token,optional"`
	Count     int32               `json:"count,optional"`
	Filters   []SearchNotesFilter `json:"filters,optional"`
}

type SearchNotesRes struct {
	Items     []*model.FeedNoteItem `json:"items"`
	NextToken string                `json:"next_token"`
	HasNext   bool                  `json:"has_next"`
	Total     int64                 `json:"total"`
}
