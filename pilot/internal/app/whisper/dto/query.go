package dto

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/errors"
)

type ListRecentChatsQuery struct {
	Uid    int64  `form:"-"`
	Cursor string `form:"cursor,optional"`
	Count  int32  `form:"count,default=30"`
}

func (q *ListRecentChatsQuery) Validate() error {
	if q == nil {
		return xerror.ErrNilArg
	}
	if q.Count <= 0 {
		q.Count = 30
	}
	if q.Count > 50 {
		q.Count = 50
	}
	return nil
}

type ListChatMsgsQuery struct {
	Uid    int64  `form:"-"`
	ChatId string `form:"chat_id"`
	Pos    int64  `form:"pos,optional"`
	Count  int32  `form:"count,default=50"`
}

func (q *ListChatMsgsQuery) Validate() error {
	if q == nil {
		return xerror.ErrNilArg
	}
	if q.ChatId == "" {
		return errors.ErrChatNotExists
	}
	if q.Count <= 0 {
		q.Count = 50
	}
	if q.Count > 100 {
		q.Count = 100
	}
	return nil
}
