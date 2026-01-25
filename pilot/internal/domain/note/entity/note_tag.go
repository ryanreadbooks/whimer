package entity

type NoteTag struct {
	Id   int64
	Name string
}

// SearchedNoteTag 搜索结果中的标签
type SearchedNoteTag struct {
	Id    string
	Name  string
	Ctime int64 // optional
}
