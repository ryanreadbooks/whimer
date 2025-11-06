package grpc

import (
	"github.com/ryanreadbooks/whimer/msger/api/msg"
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
)

func ToBizMsgContentText(t *pbuserchat.MsgReq_Text) *userchat.MsgContentText {
	return &userchat.MsgContentText{
		Text: t.Text.GetContent(),
	}
}

func ToPbMsgContentText(t *userchat.MsgContentText) *msg.MsgContentText {
	return &msg.MsgContentText{
		Content: t.Text,
	}
}

func ToBizMsgContentImage(i *pbuserchat.MsgReq_Image) *userchat.MsgContentImage {
	return &userchat.MsgContentImage{
		Key:    i.Image.Key,
		Height: i.Image.Height,
		Width:  i.Image.Width,
		Format: i.Image.Format,
	}
}

func ToPbMsgContentImage(i *userchat.MsgContentImage) *msg.MsgContentImage {
	return &msg.MsgContentImage{
		Key:    i.Key,
		Height: i.Height,
		Width:  i.Width,
		Format: i.Format,
	}
}
