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

// search related variables
var (
	// 排序规则
	NoteFilterSortRule = searchv1.NoteFilterType_SORT_RULE.String()
	// 笔记类型筛选规则
	NoteFilterNoteType = searchv1.NoteFilterType_NOTE_TYPE.String()
	// 笔记发布时间筛选规则
	NoteFilterPubTime = searchv1.NoteFilterType_NOTE_PUBTIME.String()
)

const (
	// 排序可选值
	NoteFilterSortByRelevance = "relevance"      // 综合排序（默认排序方式）
	NoteFilterSortByNewest    = "newest"         // 发布时间
	NoteFilterSortByLikes     = "likes_count"    // 点赞数量
	NoteFilterSortByComments  = "comments_count" // 评论数量

	// 笔记类型可选值
	NoteFilterNoteTypeAll   = "all"
	NoteFilterNoteTypeImage = string(NoteAssetTypeImage)

	// 发布时间筛选可选值
	NoteFilterPubTimeAll        = "all"      // 不限
	NoteFilterPubTimeInOneDay   = "oneday"   // 一天内
	NoteFilterPubTimeInOneWeek  = "oneweek"  // 一周内
	NoteFilterPubTimeInHalfYear = "halfyear" // 半年内
)
