package model

import userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"

type RecentChat struct {
	Uid         int64            `json:"-"`
	ChatId      string           `json:"chat_id"`
	ChatType    ChatType         `json:"chat_type"`
	ChatName    string           `json:"chat_name,omitempty"`
	ChatCreator int64            `json:"chat_creator,omitempty"`
	LastMsg     *Msg             `json:"last_msg,omitempty"`
	UnreadCount int64            `json:"unread_count"`
	Mtime       int64            `json:"mtime"`
	IsPinned    bool             `json:"is_pinned"`
	Cover       string           `json:"cover"`          // 单聊为对方头像 群聊为群聊头像
	Peer        *userv1.UserInfo `json:"peer,omitempty"` // 单聊对象 这里保留 userv1.UserInfo 因为来自 chat_msg.go 的 gRPC 调用
}
