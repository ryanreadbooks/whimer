package dto

import (
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

// 面向端的信息流笔记展示实体
type FeedNote struct {
	NoteId   notevo.NoteId   `json:"note_id"`
	Title    string          `json:"title"`
	Desc     string          `json:"desc"`
	Type     notevo.NoteType `json:"type"`
	CreateAt int64           `json:"create_at"`
	UpdateAt int64           `json:"update_at"`

	Images commondto.NoteImageList `json:"images"`
	Videos commondto.NoteVideoList `json:"videos"`

	Likes    int64 `json:"likes"`    // 笔记总点赞数
	Comments int64 `json:"comments"` // 笔记评论数量

	IpLoc string `json:"ip_loc"`

	Author   *commondto.User     `json:"author"`
	Interact FeedNoteInteraction `json:"interact"`
}

type FeedNoteInteraction struct {
	commondto.NoteInteraction
	Following bool `json:"following"` // 用户是否关注了笔记作者
}

type FullFeedNote struct {
	*FeedNote

	// 更多信息
	TagList []*commondto.NoteTag `json:"tag_list,omitempty"`
	AtUsers []*commondto.AtUser  `json:"at_users,omitempty"`
}
