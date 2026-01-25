package dto

type AtUser struct {
	Nickname string `json:"nickname"`
	Uid      int64  `json:"uid"`
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
