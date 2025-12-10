package model

import (
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
)

type User = userv1.UserInfo

type UserStat struct {
	Posted     int64 `json:"posted"`
	Fans       int64 `json:"fans"`
	Followings int64 `json:"followings"`
}

// 关注关系描述
type RelationStatus string

const (
	RelationFollowing RelationStatus = "following"
	RelationNone      RelationStatus = "none"
)

type HoverInfo struct {
	BasicInfo struct {
		Nickname  string `json:"nickname"`
		StyleSign string `json:"style_sign"`
		Avatar    string `json:"avatar"`
	} `json:"basic_info"`

	Interaction struct {
		Fans    string `json:"fans"`
		Follows string `json:"follows"`
	} `json:"interaction"`

	Relation struct {
		Status RelationStatus `json:"status"`
	} `json:"relation"`

	// 返回最近的发布信息
	RecentPosts []PostAsset `json:"recent_posts"`
}

type PostAsset struct {
	Url    string `json:"url"`
	Type   int    `json:"type"`
	UrlPrv string `json:"url_prv"`
}

// 模糊化处理数字
func HideActualCount(cnt int64) string {
	if cnt == 0 {
		return "0"
	}

	if cnt >= 1 && cnt < 10 {
		return "1+"
	}

	if cnt >= 10 && cnt < 100 {
		return "10+"
	}

	if cnt >= 100 && cnt < 1000 {
		return "100+"
	}

	if cnt >= 1000 && cnt < 10000 {
		return "1千+"
	}

	if cnt >= 10000 && cnt < 100000 {
		return "1万+"
	}

	if cnt >= 100000 && cnt < 1000000 {
		return "10万+"
	}

	return "100万+"
}

type UserSettings struct {
	ShowFanList    bool `json:"show_fan_list"`    // 是否展示粉丝列表
	ShowFollowList bool `json:"show_follow_list"` // 是否展示关注列表
	IntegralUserSetting
}

type IntegralNoteShowSetting struct {
	ShowNoteLikes bool `json:"show_note_likes"` // 是否公开点赞记录
}

type IntegralUserSetting struct {
	IntegralNoteShowSetting
}
