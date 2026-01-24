package vo

type Visibility int8

const (
	VisibilityPublic  Visibility = 1
	VisibilityPrivate Visibility = 2
)

func (v Visibility) CheckValid() bool {
	return v == VisibilityPublic || v == VisibilityPrivate
}

type AssetType int8

const (
	AssetTypeImage AssetType = 1
	AssetTypeVideo AssetType = 2
)

func (v AssetType) CheckValid() bool {
	return v == AssetTypeImage || v == AssetTypeVideo
}

func (v AssetType) AsNoteType() NoteType {
	switch v {
	case AssetTypeImage:
		return NoteTypeImage
	case AssetTypeVideo:
		return NoteTypeVideo
	default:
		return ""
	}
}

type NoteType string

const (
	NoteTypeImage NoteType = "image"
	NoteTypeVideo NoteType = "video"
)

func (n NoteType) CheckValid() bool {
	return n == NoteTypeImage || n == NoteTypeVideo
}

func (n NoteType) IsImage() bool {
	return n == NoteTypeImage
}

func (n NoteType) IsVideo() bool {
	return n == NoteTypeVideo
}

func (n NoteType) AsAssetType() AssetType {
	switch n {
	case NoteTypeImage:
		return AssetTypeImage
	case NoteTypeVideo:
		return AssetTypeVideo
	default:
		return AssetType(0)
	}
}

// NoteStatus 笔记状态
type NoteStatus string

const (
	NoteStatusUnknown   NoteStatus = ""
	NoteStatusPublished NoteStatus = "published"
	NoteStatusAuditing  NoteStatus = "auditing"
	NoteStatusRejected  NoteStatus = "rejected"
	NoteStatusBanned    NoteStatus = "banned"
)

func (s NoteStatus) IsValid() bool {
	switch s {
	case NoteStatusPublished, NoteStatusAuditing, NoteStatusBanned:
		return true
	}
	return false
}

type LikeAction int8

const (
	LikeActionUndo LikeAction = 0
	LikeActionDo   LikeAction = 1
)
