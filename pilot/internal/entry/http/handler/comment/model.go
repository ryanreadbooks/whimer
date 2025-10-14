package comment

import usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"

type MentionUserReq struct {
	Search string `form:"search,optional"`
}

type MentionUserGroup string

const (
	MentionRecentContacts MentionUserGroup = "recent_contacts"
	MentionFollowings     MentionUserGroup = "followings"
	MentionOthers         MentionUserGroup = "others"
)

func (g MentionUserGroup) Desc() string {
	switch g {
	case MentionRecentContacts:
		return "最近联系"
	case MentionFollowings:
		return "我的关注"
	case MentionOthers:
		return "其他"
	default:
		return ""
	}
}

type MentionUserRespItem struct {
	Group     MentionUserGroup  `json:"group"`
	GroupDesc string            `json:"group_desc"`
	Users     []*usermodel.User `json:"users"`
}

type MentionUserResp struct {
	Groups []*MentionUserRespItem `json:"groups"`
}
