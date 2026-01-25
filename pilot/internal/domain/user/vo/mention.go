package vo

// @用户分组类型
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

// @用户响应项
type MentionUserItem struct {
	Group     MentionUserGroup
	GroupDesc string
	Users     []*User
}
