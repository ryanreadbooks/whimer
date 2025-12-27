package model

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
	Group     MentionUserGroup `json:"group"`
	GroupDesc string           `json:"group_desc"`
	Users     []*User          `json:"users"`
}
