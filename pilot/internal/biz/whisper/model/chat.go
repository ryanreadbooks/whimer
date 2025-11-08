package model

import userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"

type ChatType string

const (
	P2PChat   ChatType = "p2p"
	GroupChat ChatType = "group"
)

func ChatTypeFromPb(t userchatv1.ChatType) ChatType {
	switch t {
	case userchatv1.ChatType_P2P:
		return P2PChat
	case userchatv1.ChatType_GROUP:
		return GroupChat
	}

	return ""
}

var (
	validChatType = map[ChatType]struct{}{
		P2PChat:   {},
		GroupChat: {},
	}
)

func IsValidChatType(s string) bool {
	_, ok := validChatType[ChatType(s)]
	return ok
}

type Chat struct {
	Name  string   `json:"name,omitempty"`
	Type  ChatType `json:"type"`
	Ctime int64    `json:"ctime"`
}
