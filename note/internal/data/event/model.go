package event

import (
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/model/event"
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

	return &event.Note{
		Id:      note.NoteId,
		Title:   note.Title,
		Desc:    note.Desc,
		Type:    note.Type.String(),
		Owner:   note.Owner,
		Ctime:   note.CreateAt,
		Utime:   note.UpdateAt,
		Ip:      note.Ip,
		Images:  images,
		Videos:  videos,
		Tags:    note.Tags,
		AtUsers: note.AtUsers,
	}
}
