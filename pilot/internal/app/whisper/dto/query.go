package dto

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/errors"
)

type Order string

const (
	OrderAsc  Order = "asc"
	OrderDesc Order = "desc"
)

func (o Order) Valid() bool {
	return o == OrderAsc || o == OrderDesc
}

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
	Order  Order  `form:"order,default=desc"`
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

	if q.Order == "" {
		q.Order = OrderDesc
	}

	if !q.Order.Valid() {
		return errors.ErrInvalidOrder
	}

	return nil
}
