package convert

import (
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

func MsgTypeToPb(t vo.MsgType) pbmsg.MsgType {
	switch t {
	case vo.MsgText:
		return pbmsg.MsgType_MSG_TYPE_TEXT
	case vo.MsgImage:
		return pbmsg.MsgType_MSG_TYPE_IMAGE
	default:
		return pbmsg.MsgType_MSG_TYPE_UNSPECIFIED
	}
}

func MsgTypeFromPb(t pbmsg.MsgType) vo.MsgType {
	switch t {
	case pbmsg.MsgType_MSG_TYPE_TEXT:
		return vo.MsgText
	case pbmsg.MsgType_MSG_TYPE_IMAGE:
		return vo.MsgImage
	default:
		return vo.MsgTypeUnspecified
	}
}

func MsgStatusFromPb(s pbmsg.MsgStatus) vo.MsgStatus {
	switch s {
	case pbmsg.MsgStatus_MSG_STATUS_NORMAL:
		return vo.MsgStatusNormal
	case pbmsg.MsgStatus_MSG_STATUS_RECALL:
		return vo.MsgStatusRecall
	default:
		return vo.MsgStatusNormal
	}
}

func SendMsgParamsToPb(params *repository.SendMsgParams) *userchatv1.MsgReq {
	pbReq := &userchatv1.MsgReq{
		Cid:  params.Cid,
		Type: MsgTypeToPb(params.Type),
	}

	switch params.Type {
	case vo.MsgText:
		if params.Content != nil && params.Content.Text != nil {
			pbReq.Content = &userchatv1.MsgReq_Text{
				Text: &pbmsg.MsgContentText{Content: params.Content.Text.Content},
			}
		}
	case vo.MsgImage:
		if params.Content != nil && params.Content.Image != nil {
			pbReq.Content = &userchatv1.MsgReq_Image{
				Image: &pbmsg.MsgContentImage{
					Key:    params.Content.Image.Key,
					Height: params.Content.Image.Height,
					Width:  params.Content.Image.Width,
					Format: params.Content.Image.Format,
				},
			}
		}
	}

	return pbReq
}

func ChatTypeFromPb(t userchatv1.ChatType) vo.ChatType {
	switch t {
	case userchatv1.ChatType_P2P:
		return vo.P2PChat
	case userchatv1.ChatType_GROUP:
		return vo.GroupChat
	}
	return ""
}

func RecentChatFromPb(pb *userchatv1.RecentChat) *entity.RecentChat {
	return &entity.RecentChat{
		Uid:         pb.Uid,
		ChatId:      pb.ChatId,
		ChatType:    ChatTypeFromPb(pb.ChatType),
		ChatName:    pb.ChatName,
		ChatCreator: pb.ChatCreator,
		LastMsg:     MsgFromPb(pb.GetLastMsg()),
		UnreadCount: pb.UnreadCount,
		Mtime:       pb.Mtime,
		IsPinned:    pb.IsPinned,
	}
}

func MsgFromPb(pbChatMsg *userchatv1.ChatMsg) *entity.Msg {
	if pbChatMsg == nil {
		return nil
	}

	pbMsg := pbChatMsg.GetMsg()
	if pbMsg == nil {
		return nil
	}

	msg := &entity.Msg{
		Id:        pbMsg.GetId(),
		Type:      MsgTypeFromPb(pbMsg.GetType()),
		Cid:       pbMsg.GetCid(),
		Status:    MsgStatusFromPb(pbMsg.GetStatus()),
		Mtime:     pbMsg.GetMtime(),
		SenderUid: pbMsg.GetSender(),
		Pos:       pbChatMsg.GetPos(),
		Ext:       MsgExtFromPb(pbMsg.GetExt()),
	}

	if msg.Id != "" && msg.Status != vo.MsgStatusRecall {
		msg.Content = MsgContentFromPb(msg.Type, pbMsg)
	}

	return msg
}

func MsgExtFromPb(pbext *pbmsg.MsgExt) *vo.MsgExt {
	if pbext == nil {
		return nil
	}
	ext := &vo.MsgExt{}
	if pbext.Recall != nil {
		ext.Recall = &vo.MsgExtRecall{
			RecallUid: pbext.Recall.GetUid(),
			RecallAt:  pbext.Recall.GetTime(),
		}
	}
	return ext
}

func MsgContentFromPb(msgType vo.MsgType, pb *pbmsg.Msg) *vo.MsgContent {
	content := &vo.MsgContent{ContentType: msgType}

	switch msgType {
	case vo.MsgText:
		content.Text = &vo.MsgTextContent{
			Content: pb.GetText().GetContent(),
			Preview: pb.GetText().GetPreview(),
		}
	case vo.MsgImage:
		content.Image = &vo.MsgImageContent{
			Key:    pb.GetImage().GetKey(),
			Height: pb.GetImage().GetHeight(),
			Width:  pb.GetImage().GetWidth(),
			Format: pb.GetImage().GetFormat(),
		}
	}

	return content
}
