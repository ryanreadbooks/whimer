package entity

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

type Msg struct {
	Id        string
	Cid       string
	Type      vo.MsgType
	Status    vo.MsgStatus
	Mtime     int64
	SenderUid int64
	Content   *vo.MsgContent
	Pos       int64
	Ext       *vo.MsgExt
}
