package dto

import (
	"strings"

	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator/errors"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// PageListNotesQuery 分页查询笔记请求
type PageListNotesQuery struct {
	Page   int32             `form:"page,optional"`
	Count  int32             `form:"count,default=15"`
	Status notevo.NoteStatus `form:"status,default=published"`
}

func (q *PageListNotesQuery) Validate() error {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.Count <= 0 {
		q.Count = 15
	}
	if q.Count > 50 {
		q.Count = 50
	}
	if !q.Status.IsValid() {
		q.Status = notevo.NoteStatusPublished
	}
	return nil
}

// Note 创作者笔记详情结果
//
// 面向端
type Note struct {
	NoteId   notevo.NoteId   `json:"note_id"`
	Title    string          `json:"title"`
	Desc     string          `json:"desc"`
	Privacy  int8            `json:"privacy"`
	CreateAt int64           `json:"create_at"`
	UpdateAt int64           `json:"update_at"`
	IpLoc    string          `json:"ip_loc"`
	Type     notevo.NoteType `json:"type"`

	Images commondto.NoteImageList `json:"images,omitempty"`
	Videos commondto.NoteVideoList `json:"videos,omitempty"`

	Likes   int64 `json:"likes"`
	Replies int64 `json:"replies"`

	AtUsers []*commondto.AtUser  `json:"at_users,omitempty"`
	TagList []*commondto.NoteTag `json:"tag_list,omitempty"`

	Interact commondto.NoteInteraction `json:"interact"`
}

type NoteList struct {
	Total int64   `json:"total"`
	Items []*Note `json:"items"`
}

// SearchedTag 搜索到的标签
type SearchedTag struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type SearchTagsQuery struct {
	Name string `json:"name"`
}

func (r *SearchTagsQuery) Validate() error {
	if r == nil {
		return errors.ErrNilArg
	}

	r.Name = strings.TrimSpace(r.Name)
	if r.Name == "" {
		return nil
	}

	if err := checkTagName(r.Name); err != nil {
		return err
	}

	return nil
}
