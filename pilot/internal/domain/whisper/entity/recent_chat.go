package entity

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

type RecentChat struct {
	Uid         int64
	ChatId      string
	ChatType    vo.ChatType
	ChatName    string
	ChatCreator int64
	LastMsg     *Msg
	UnreadCount int64
	Mtime       int64
	IsPinned    bool
	Cover       string
	PeerUid     int64
}
