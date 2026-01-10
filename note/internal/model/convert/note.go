package convert

import (
	"github.com/ryanreadbooks/whimer/misc/xnet"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	tagdao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/tag"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// 只有基础字段 不包含image等需要额外设置的字段
func NoteFromDao(d *notedao.NotePO) *model.Note {
	n := &model.Note{}
	if d == nil {
		return n
	}
	n.NoteId = d.Id
	n.Title = d.Title
	n.Desc = d.Desc
	n.Privacy = d.Privacy
	n.Type = d.NoteType
	n.State = d.State
	n.CreateAt = d.CreateAt
	n.UpdateAt = d.UpdateAt
	n.Ip = xnet.BytesIpAsString(d.Ip)
	n.Owner = d.Owner

	return n
}

func NoteCoreFromDao(d *notedao.NotePO) *model.NoteCore {
	n := &model.NoteCore{}
	if d == nil {
		return n
	}
	n.NoteId = d.Id
	n.Title = d.Title
	n.Desc = d.Desc
	n.Privacy = d.Privacy
	n.Type = d.NoteType
	n.State = d.State
	n.CreateAt = d.CreateAt
	n.UpdateAt = d.UpdateAt
	n.Ip = xnet.BytesIpAsString(d.Ip)
	n.Owner = d.Owner
	return n
}

func NoteSliceFromDao(ds []*notedao.NotePO) []*model.Note {
	notes := make([]*model.Note, 0, len(ds))
	for _, n := range ds {
		notes = append(notes, NoteFromDao(n))
	}
	return notes
}

func NoteTagFromDao(d *tagdao.Tag) *model.NoteTag {
	return &model.NoteTag{
		Id:    d.Id,
		Name:  d.Name,
		Ctime: d.Ctime,
	}
}
