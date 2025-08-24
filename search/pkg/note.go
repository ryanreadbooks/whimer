package pkg

import searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

// 对外常量定义

type NoteAssetType string

const (
	NoteAssetTypeImage NoteAssetType = "image"
	// NoteAssetTypeVideo NoteAssetType = "video"
)

var (
	NoteAssetConverter = map[searchv1.Note_AssetType]NoteAssetType{
		searchv1.Note_ASSET_TYPE_UNSPECIFIED: "all",
		searchv1.Note_ASSET_TYPE_IMAGE:       NoteAssetTypeImage,
	}
)

type NoteVisibility string

const (
	VisibilityPublic  NoteVisibility = "public"
	VisibilityPrivate NoteVisibility = "private"
)

var (
	NoteVisibilityConverter = map[searchv1.Note_Visibility]NoteVisibility{
		searchv1.Note_VISIBILITY_PUBLIC:  VisibilityPublic,
		searchv1.Note_VISIBILITY_PRIVATE: VisibilityPrivate,
	}
)

// search related constants
