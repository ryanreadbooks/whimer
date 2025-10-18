package model

import (
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type NoteExt struct {
	TagIds  []int64
	AtUsers []*AtUser
}

func (e *NoteExt) SetTagIds(s string) {
	e.TagIds = xslice.SplitInts[int64](s, ",")
}

type NoteTag struct {
	Id    int64  `json:"id"`
	Name  string `json:"name"`
	Ctime int64  `json:"ctime"`
}

func (t *NoteTag) AsPb() *notev1.NoteTag {
	if t == nil {
		return nil
	}

	return &notev1.NoteTag{
		Id:    t.Id,
		Name:  t.Name,
		Ctime: t.Ctime,
	}
}

func NoteTagListAsPb(tags []*NoteTag) []*notev1.NoteTag {
	if len(tags) == 0 {
		return nil
	}

	var r []*notev1.NoteTag
	for _, t := range tags {
		if t == nil {
			continue
		}
		r = append(r, t.AsPb())
	}

	return r
}

type AtUser struct {
	Nickname string `json:"nickname"`
	Uid      int64  `json:"uid"`
}

func (a *AtUser) AsPb() *notev1.NoteAtUser {
	if a == nil {
		return nil
	}

	return &notev1.NoteAtUser{
		Nickname: a.Nickname,
		Uid:      a.Uid,
	}
}

func AtUsersAsPb(atUsers []*AtUser) []*notev1.NoteAtUser {
	if len(atUsers) == 0 {
		return nil
	}

	var r []*notev1.NoteAtUser
	for _, a := range atUsers {
		if a == nil {
			continue
		}
		r = append(r, a.AsPb())
	}

	return r
}

func AtUsersFromPb(atUsers []*notev1.NoteAtUser) []*AtUser {
	if len(atUsers) == 0 {
		return nil
	}

	var r []*AtUser
	for _, a := range atUsers {
		r = append(r, &AtUser{
			Nickname: a.Nickname,
			Uid:      a.Uid,
		})
	}

	return r
}

func FilterInvalidAtUsers(atUsers []*AtUser) []*AtUser {
	if len(atUsers) == 0 {
		return nil
	}

	var r []*AtUser
	for _, a := range atUsers {
		if len(a.Nickname) == 0 || a.Uid == 0 {
			continue
		}

		r = append(r, a)
	}

	return r
}
