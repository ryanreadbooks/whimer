package creator

type ListResItemImage struct {
	Url  string `json:"url"`
	Type int    `json:"type"`
}

type ListResItem struct {
	NoteId   string              `json:"note_id"`
	Title    string              `json:"title"`
	Desc     string              `json:"desc"`
	Privacy  int64               `json:"privacy"`
	CreateAt int64               `json:"create_at"`
	UpdateAt int64               `json:"update_at"`
	Images   []*ListResItemImage `json:"images"`
}

type ListRes struct {
	Items []*ListResItem `json:"items"`
}

type GetNoteReq struct {
	NoteId string `path:"note_id"`
}
