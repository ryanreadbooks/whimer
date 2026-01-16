package event

import (
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/model/event"
	pkgid "github.com/ryanreadbooks/whimer/note/pkg/id"
)

func modelNoteToEventNote(note *model.Note) *event.Note {
	images := make([]string, 0, len(note.Images))
	for _, image := range note.Images {
		images = append(images, image.Key)
	}

	var videos []string
	if note.Videos != nil {
		videos = make([]string, 0, len(note.Videos.Items))
		for _, video := range note.Videos.Items {
			videos = append(videos, video.Key)
		}
	}

	noteTags := make([]*event.NoteTag, 0, len(note.Tags))
	for _, tag := range note.Tags {
		noteTags = append(noteTags, &event.NoteTag{
			Id:    tag.Id,
			Tid:   pkgid.TagId(tag.Id).String(),
			Name:  tag.Name,
			Ctime: tag.Ctime,
		})
	}

	return &event.Note{
		Id:      note.NoteId,
		Nid:     pkgid.NoteId(note.NoteId).String(),
		Title:   note.Title,
		Desc:    note.Desc,
		Type:    note.Type.String(),
		Owner:   note.Owner,
		Ctime:   note.CreateAt,
		Utime:   note.UpdateAt,
		Ip:      note.Ip,
		Images:  images,
		Videos:  videos,
		Tags:    noteTags,
		AtUsers: note.AtUsers,
	}
}
