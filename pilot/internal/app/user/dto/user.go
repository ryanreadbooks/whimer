package dto

import (
	"strconv"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
)

type User struct {
	Uid       int64  `json:"uid"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	StyleSign string `json:"style_sign"`
	Gender    string `json:"gender"`
}

func ConvertVoUserToDto(u *vo.User) *User {
	if u == nil {
		return nil
	}
	return &User{
		Uid:       u.Uid,
		Nickname:  u.Nickname,
		Avatar:    u.Avatar,
		StyleSign: u.StyleSign,
		Gender:    u.Gender,
	}
}

type UserStat struct {
	Posted     int64 `json:"posted"`
	Fans       int64 `json:"fans"`
	Followings int64 `json:"followings"`
}

func ConvertVoUserStatToDto(s *vo.UserStat) *UserStat {
	if s == nil {
		return nil
	}
	return &UserStat{
		Posted:     s.Posted,
		Fans:       s.Fans,
		Followings: s.Followings,
	}
}

type UserWithRelation struct {
	User     *User  `json:"user"`
	Relation string `json:"relation"`
}

func ConvertVoUserWithRelationToDto(u *vo.UserWithRelation) *UserWithRelation {
	if u == nil {
		return nil
	}
	return &UserWithRelation{
		User:     ConvertVoUserToDto(u.User),
		Relation: string(u.Relation),
	}
}

type FanOrFollowingListResult struct {
	Items []*UserWithRelation `json:"items"`
	Total int64               `json:"total"`
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
		Status string `json:"status"`
	} `json:"relation"`

	RecentPosts []RecentPost `json:"recent_posts"`
}

type RecentPost struct {
	NoteId string `json:"note_id"`
	Type   string `json:"type"`
	Cover  string `json:"cover"`
}

func ConvertVoHoverInfoToDto(h *vo.HoverInfo, isAuthed bool) *HoverInfo {
	if h == nil {
		return nil
	}

	result := &HoverInfo{}
	result.BasicInfo.Nickname = h.BasicInfo.Nickname
	result.BasicInfo.StyleSign = h.BasicInfo.StyleSign
	result.BasicInfo.Avatar = h.BasicInfo.Avatar
	result.Relation.Status = string(h.Relation.Status)

	// 非登录用户不展示准确数值
	if !isAuthed {
		fansCount, _ := strconv.ParseInt(h.Interaction.Fans, 10, 64)
		followsCount, _ := strconv.ParseInt(h.Interaction.Follows, 10, 64)
		result.Interaction.Fans = HideActualCount(fansCount)
		result.Interaction.Follows = HideActualCount(followsCount)
	} else {
		result.Interaction.Fans = h.Interaction.Fans
		result.Interaction.Follows = h.Interaction.Follows
	}

	result.RecentPosts = make([]RecentPost, 0, len(h.RecentPosts))
	for _, p := range h.RecentPosts {
		result.RecentPosts = append(result.RecentPosts, RecentPost{
			NoteId: p.NoteId.String(),
			Type:   string(p.Type),
			Cover:  p.Cover,
		})
	}

	return result
}

type MentionUserItem struct {
	Group     string  `json:"group"`
	GroupDesc string  `json:"group_desc"`
	Users     []*User `json:"users"`
}

func ConvertVoMentionItemToDto(item *vo.MentionUserItem) *MentionUserItem {
	if item == nil {
		return nil
	}
	users := make([]*User, 0, len(item.Users))
	for _, u := range item.Users {
		users = append(users, ConvertVoUserToDto(u))
	}
	return &MentionUserItem{
		Group:     string(item.Group),
		GroupDesc: item.GroupDesc,
		Users:     users,
	}
}

type UserSettings struct {
	ShowFanList    bool `json:"show_fan_list"`
	ShowFollowList bool `json:"show_follow_list"`
	ShowNoteLikes  bool `json:"show_note_likes"`
}
