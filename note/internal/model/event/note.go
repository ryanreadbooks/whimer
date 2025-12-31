package event

import "github.com/ryanreadbooks/whimer/note/internal/model"

type Note struct {
	Id      int64            `json:"id"`
	Title   string           `json:"title"`
	Desc    string           `json:"desc"`
	Type    string           `json:"type"`
	Owner   int64            `json:"owner"`
	Ctime   int64            `json:"create_at"`
	Utime   int64            `json:"update_at"`
	Ip      string           `json:"ip"`
	Images  []string         `json:"images"` // asset keys
	Videos  []string         `json:"videos"` // asset keys
	Tags    []*model.NoteTag `json:"tags,omitempty"`
	AtUsers []*model.AtUser  `json:"at_users,omitempty"`
}
