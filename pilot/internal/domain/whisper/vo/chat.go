package vo

type ChatType string

const (
	P2PChat   ChatType = "p2p"
	GroupChat ChatType = "group"
)

var validChatType = map[ChatType]struct{}{
	P2PChat:   {},
	GroupChat: {},
}

func IsValidChatType(s string) bool {
	_, ok := validChatType[ChatType(s)]
	return ok
}
