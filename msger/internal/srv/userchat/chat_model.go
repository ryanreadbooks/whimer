package userchat

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type SendMsgReq struct {
	Type  model.MsgType
	Text  *userchat.MsgContentText
	Image *userchat.MsgContentImage
	Cid   string

	content []byte // need to be filled explicitly
}

func (c *SendMsgReq) FillContent() error {
	if c == nil {
		return global.ErrArgs.Msg("send msg req nil")
	}

	var ic userchat.MsgContent
	switch c.Type {
	case model.MsgText:
		ic = c.Text
	case model.MsgImage:
		ic = c.Image
	default:
		return global.ErrUnsupportedMsgType
	}

	content, err := ic.Bytes()
	if err != nil {
		return err
	}

	c.content = content

	return nil
}
