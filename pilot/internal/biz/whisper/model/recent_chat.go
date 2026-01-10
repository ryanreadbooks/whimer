package model

import usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user/model"

type RecentChat struct {
	Uid         int64           `json:"-"`
	ChatId      string          `json:"chat_id"`
	ChatType    ChatType        `json:"chat_type"`
	ChatName    string          `json:"chat_name,omitempty"`
	ChatCreator int64           `json:"chat_creator,omitempty"`
	LastMsg     *Msg            `json:"last_msg,omitempty"`
	UnreadCount int64           `json:"unread_count"`
	Mtime       int64           `json:"mtime"`
	IsPinned    bool            `json:"is_pinned"`
	Cover       string          `json:"cover"`          // 单聊为对方头像 群聊为群聊头像
	Peer        *usermodel.User `json:"peer,omitempty"` // 单聊对象
}
