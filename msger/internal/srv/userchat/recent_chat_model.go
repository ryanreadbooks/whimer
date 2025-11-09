package userchat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	biz "github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type RecentChat struct {
	Uid           int64
	ChatId        uuid.UUID
	ChatType      model.ChatType
	ChatName      string
	ChatStatus    model.ChatStatus
	ChatCreator   int64
	LastMsg       *biz.Msg
	LastReadMsgId uuid.UUID
	LastReadTime  int64
	UnreadCount   int64
	Ctime         int64
	Mtime         int64
	IsPinned      bool
}
