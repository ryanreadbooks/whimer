package event

import "github.com/ryanreadbooks/whimer/note/internal/model"

type Note struct {
	Id      int64           `json:"id"`
	Nid     string          `json:"nid"` // 字符串格式的Id
	Title   string          `json:"title"`
	Desc    string          `json:"desc"`
	Type    string          `json:"type"`
	Owner   int64           `json:"owner"`
	Ctime   int64           `json:"create_at"`
	Utime   int64           `json:"update_at"`
	Ip      string          `json:"ip"`
	Images  []string        `json:"images,omitempty"` // asset keys
	Videos  []string        `json:"videos,omitempty"` // asset keys
	Tags    []*NoteTag      `json:"tags,omitempty"`
	AtUsers []*model.AtUser `json:"at_users,omitempty"`
}

type NoteTag struct {
	Id    int64  `json:"id"`
	Tid   string `json:"tid"` // 字符串格式的Id
	Name  string `json:"name"`
	Ctime int64  `json:"ctime"`
}
