package model

type AtUser struct {
	Nickname string `json:"nickname"`
	Uid      int64  `json:"uid"`
}

func (a *AtUser) GetNickname() string {
	if a != nil {
		return a.Nickname
	}
	return ""
}

func (a *AtUser) GetUid() int64 {
	if a != nil {
		return a.Uid
	}
	return 0
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

type IAtUser interface {
	GetNickname() string
	GetUid() int64
}
