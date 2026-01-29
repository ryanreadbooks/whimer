package convert

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

func EntitySearchNoteToPb(note *entity.SearchNote) *searchv1.Note {
	return &searchv1.Note{
		NoteId:   note.NoteId.String(),
		Title:    note.Title,
		Desc:     note.Desc,
		CreateAt: note.CreateAt,
		UpdateAt: note.UpdateAt,
		Author: &searchv1.Note_Author{
			Uid:      note.AuthorUid,
			Nickname: note.AuthorNickname,
		},
		TagList:    BatchEntitySearchedNoteTagsAsPb(note.TagList),
		AssetType:  VoAssetTypeAsPb(note.AssetType),
		Visibility: VoVisibilityAsPb(note.Visibility),
	}
}

func EntitySearchedNoteTagAsPb(tag *entity.SearchedNoteTag) *searchv1.NoteTag {
	return &searchv1.NoteTag{
		Id:    tag.Id,
		Name:  tag.Name,
		Ctime: tag.Ctime,
	}
}

func BatchEntitySearchedNoteTagsAsPb(tags []*entity.SearchedNoteTag) []*searchv1.NoteTag {
	if len(tags) == 0 {
		return nil
	}

	tagList := make([]*searchv1.NoteTag, 0, len(tags))
	for _, tag := range tags {
		tagList = append(tagList, EntitySearchedNoteTagAsPb(tag))
	}
	return tagList
}

func VoAssetTypeAsPb(assetType vo.AssetType) searchv1.Note_AssetType {
	switch assetType {
	case vo.AssetTypeImage:
		return searchv1.Note_ASSET_TYPE_IMAGE
	case vo.AssetTypeVideo:
		return searchv1.Note_ASSET_TYPE_VIDEO
	}
	return searchv1.Note_ASSET_TYPE_UNSPECIFIED
}

func VoVisibilityAsPb(visibility vo.Visibility) searchv1.Note_Visibility {
	switch visibility {
	case vo.VisibilityPublic:
		return searchv1.Note_VISIBILITY_PUBLIC
	case vo.VisibilityPrivate:
		return searchv1.Note_VISIBILITY_PRIVATE
	}
	return searchv1.Note_VISIBILITY_UNSPECIFIED
}
