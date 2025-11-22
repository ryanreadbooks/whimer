package grpc

import (
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	bizuserchat "github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv/userchat"
	userchatsrv "github.com/ryanreadbooks/whimer/msger/internal/srv/userchat"
)

func ToPbMsgExt(ext *bizuserchat.MsgExt) *pbmsg.MsgExt {
	if ext == nil {
		return &pbmsg.MsgExt{}
	}

	return &pbmsg.MsgExt{
		Recall: &pbmsg.MsgExtRecall{
			Uid:  ext.Recall.GetUid(),
			Time: ext.Recall.GetTime(),
		},
	}
}

func BindPbMsgContent(msg *bizuserchat.Msg, pb *pbmsg.Msg) {
	switch msg.Type {
	case model.MsgText:
		text, ok := msg.Content.(*bizuserchat.MsgContentText)
		pb.Text = &pbmsg.MsgContentText{}
		if ok {
			pb.Text.Content = text.Text
			pb.Text.Preview = text.Preview()
		}
	case model.MsgImage:
		img, ok := msg.Content.(*bizuserchat.MsgContentImage)
		pb.Image = &pbmsg.MsgContentImage{}
		if ok {
			pb.Image.Key = img.Key
			pb.Image.Height = img.Height
			pb.Image.Width = img.Width
			pb.Image.Format = img.Format
			pb.Image.PreviewText = img.Preview()
		}
	default:
		// msgId = 0x0
	}
}

func ToPbMsg(msg *bizuserchat.Msg) *pbmsg.Msg {
	pb := &pbmsg.Msg{
		Id:     msg.Id.String(),
		Type:   model.MsgTypeToPb(msg.Type),
		Status: model.MsgStatusToPb(msg.Status),
		Sender: msg.Sender,
		Mtime:  msg.Mtime,
		Cid:    msg.Cid,
		Ext:    ToPbMsgExt(msg.Ext),
	}

	BindPbMsgContent(msg, pb)

	return pb
}

func ToPbChatMsg(cmsg *userchat.ChatMsg) *pbuserchat.ChatMsg {
	pb := &pbuserchat.ChatMsg{
		Msg:    ToPbMsg(cmsg.Msg),
		ChatId: cmsg.ChatId.String(),
		Pos:    cmsg.Pos,
	}

	return pb
}

func ToPbChatMsgs(cmsgs []*userchat.ChatMsg) []*pbuserchat.ChatMsg {
	pbmsgs := make([]*pbuserchat.ChatMsg, 0, len(cmsgs))
	for _, m := range cmsgs {
		pbmsgs = append(pbmsgs, ToPbChatMsg(m))
	}
	return pbmsgs
}

func ToBizMsgContentText(t *pbuserchat.MsgReq_Text) *bizuserchat.MsgContentText {
	return &bizuserchat.MsgContentText{
		Text: t.Text.GetContent(),
	}
}

func ToPbMsgContentText(t *bizuserchat.MsgContentText) *pbmsg.MsgContentText {
	return &pbmsg.MsgContentText{
		Content: t.Text,
	}
}

func ToBizMsgContentImage(i *pbuserchat.MsgReq_Image) *bizuserchat.MsgContentImage {
	return &bizuserchat.MsgContentImage{
		Key:    i.Image.Key,
		Height: i.Image.Height,
		Width:  i.Image.Width,
		Format: i.Image.Format,
	}
}

func ToPbMsgContentImage(i *bizuserchat.MsgContentImage) *pbmsg.MsgContentImage {
	return &pbmsg.MsgContentImage{
		Key:    i.Key,
		Height: i.Height,
		Width:  i.Width,
		Format: i.Format,
	}
}

func ToPbRecentChat(rc *userchatsrv.RecentChat) *pbuserchat.RecentChat {
	pbrc := &pbuserchat.RecentChat{
		Uid:           rc.Uid,
		ChatId:        rc.ChatId.String(),
		ChatType:      model.ChatTypeToPb(rc.ChatType),
		ChatName:      rc.ChatName,
		ChatStatus:    model.ChatStatusToPb(rc.ChatStatus),
		ChatCreator:   rc.ChatCreator,
		LastMsg:       ToPbChatMsg(rc.LastMsg),
		LastReadMsgId: rc.LastReadMsgId.String(),
		LastReadTime:  rc.LastReadTime,
		UnreadCount:   rc.UnreadCount,
		Ctime:         rc.Ctime,
		Mtime:         rc.Mtime,
		IsPinned:      rc.IsPinned,
	}

	return pbrc
}
