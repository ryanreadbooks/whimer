package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatPO struct {
	Id        uuid.UUID        `db:"id"`
	Type      model.ChatType   `db:"type"`
	Name      string           `db:"name"`
	Status    model.ChatStatus `db:"status"`
	Creator   int64            `db:"creator"`
	Mtime     int64            `db:"mtime"`
	LastMsgId uuid.UUID        `db:"last_msg_id"`
	Settings  int64            `db:"settings"`
}
