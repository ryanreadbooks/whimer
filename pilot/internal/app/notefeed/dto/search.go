package dto

import "github.com/ryanreadbooks/whimer/misc/xerror"

// 搜索笔记过滤器
type SearchNotesFilter struct {
	Type  string `json:"type,optional"`
	Value string `json:"value,optional"`
}

// 搜索笔记请求
type SearchNotesQuery struct {
	Keyword   string              `json:"keyword"`
	PageToken string              `json:"page_token,optional"`
	Count     int32               `json:"count,optional"`
	Filters   []SearchNotesFilter `json:"filters,optional"`
}

func (r *SearchNotesQuery) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Keyword == "" {
		return xerror.ErrArgs.Msg("keyword is required")
	}

	if r.Count > 50 {
		r.Count = 50
	}
	if r.Count <= 0 {
		r.Count = 20
	}

	return nil
}

// 搜索笔记结果
type SearchNotesResult struct {
	Items     []*FeedNote `json:"items"`
	NextToken string      `json:"next_token"`
	HasNext   bool        `json:"has_next"`
	Total     int64       `json:"total"`
}
