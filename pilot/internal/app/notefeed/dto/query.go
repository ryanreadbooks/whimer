package dto

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed/errors"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

type GetRandomQuery struct {
	NeedNum  int    `form:"need_num"`
	Platform string `form:"platform,optional"`
	Category string `form:"category,optional"`
}

func (r *GetRandomQuery) Validate() error {
	const (
		maxAllowedNeedNum = 20
	)

	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NeedNum > maxAllowedNeedNum {
		return errors.ErrTooManyNotes
	}

	return nil
}

type GetFeedNoteQuery struct {
	NoteId notevo.NoteId `form:"note_id" path:"note_id"`
	Source string        `form:"source,optional"`
}

func (r *GetFeedNoteQuery) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.NoteId == 0 {
		return errors.ErrNoteNotFound
	}

	return nil
}

type ListUserFeedNotesQuery struct {
	Uid    int64         `form:"uid"`
	Cursor notevo.NoteId `form:"cursor,optional"`
	Count  int32         `form:"count,optional"`
}

func (r *ListUserFeedNotesQuery) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	if r.Count > 20 {
		r.Count = 20
	}
	if r.Count <= 0 {
		r.Count = 10
	}

	return nil
}

type ListUserFeedNotesResult struct {
	Items      []*FeedNote   `json:"items"`
	NextCursor notevo.NoteId `json:"next_cursor"`
	HasNext    bool          `json:"has_next"`
}

type ListUserLikedNoteQuery struct {
	Uid    int64  `form:"uid"`
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,optional"`
}

func (r *ListUserLikedNoteQuery) Validate() error {
	if r == nil {
		return xerror.ErrNilArg
	}

	if r.Uid == 0 {
		return errors.ErrUserNotFound
	}

	if r.Count > 20 {
		r.Count = 20
	}
	if r.Count <= 0 {
		r.Count = 10
	}

	return nil
}

type ListUserLikedNoteResult struct {
	Items      []*FeedNote `json:"items"`
	NextCursor string      `json:"next_cursor"`
	HasNext    bool        `json:"has_next"`
}
