package dto

import mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"

type AtUser struct {
	Nickname string `json:"nickname"`
	Uid      int64  `json:"uid"`
}

func (a *AtUser) ToMentionVo() *mentionvo.AtUser {
	return &mentionvo.AtUser{
		Uid:      a.Uid,
		Nickname: a.Nickname,
	}
}

type AtUserList []AtUser

func (a AtUserList) Filter() AtUserList {
	filtered := make(AtUserList, 0, len(a))
	for _, atUser := range a {
		if atUser.Uid != 0 && len(atUser.Nickname) > 0 {
			filtered = append(filtered, atUser)
		}
	}
	return filtered
}

func (a AtUserList) ToMentionVo() mentionvo.AtUserList {
	vo := make(mentionvo.AtUserList, 0, len(a))
	for _, atUser := range a {
		vo = append(vo, atUser.ToMentionVo())
	}
	return vo
}

type User struct {
	Uid       int64  `json:"uid"`
	Nickname  string `json:"nickname"`
	Avatar    string `json:"avatar"`
	StyleSign string `json:"style_sign,omitempty"`
}
