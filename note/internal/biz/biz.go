package biz

type Biz struct {
	Note     NoteBiz
	Interact NoteInteractBiz
	Creator  NoteCreatorBiz
}

func New() Biz {
	note := NewNoteBiz()
	creator := NewNoteCreatorBiz()
	interact := NewNoteInteractBiz()
	return Biz{
		Note:     note,
		Interact: interact,
		Creator:  creator,
	}
}
