package entity

type NoteTag struct {
	Id   int64
	Name string
}

// SearchedNoteTag 搜索结果中的标签（ID 为字符串格式）
type SearchedNoteTag struct {
	Id   string
	Name string
}
