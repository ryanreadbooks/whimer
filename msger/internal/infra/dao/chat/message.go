package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

// message
type MsgPO struct {
	Id uuid.UUID `db:"id"`
	Type model.MsgType `db:"type"`
}