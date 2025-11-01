package srv

import "github.com/ryanreadbooks/whimer/msger/internal/model"

type SendMsgReq struct {
	Type    model.MsgType
	Content []byte
	Cid     string
}
