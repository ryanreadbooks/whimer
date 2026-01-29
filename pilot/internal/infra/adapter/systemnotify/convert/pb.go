package convert

import (
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

func ChatUnreadFromPb(pb *systemv1.ChatUnread) entity.ChatUnread {
	return entity.ChatUnread{
		ChatId: pb.GetChatId(),
		Count:  pb.GetUnreadCount(),
	}
}

func ListMsgResultFromPb(msgs []*systemv1.SystemMsg, chatId string, hasMore bool) *vo.ListMsgResult {
	result := &vo.ListMsgResult{
		Messages: make([]*vo.RawSystemMsg, 0, len(msgs)),
		ChatId:   chatId,
		HasMore:  hasMore,
	}

	for _, msg := range msgs {
		result.Messages = append(result.Messages, &vo.RawSystemMsg{
			Id:      msg.GetId(),
			RecvUid: msg.GetRecvUid(),
			Content: msg.GetContent(),
			Status:  msgStatusFromPb(msg.GetStatus()),
		})
	}

	return result
}

func msgStatusFromPb(status systemv1.SystemMsgStatus) vo.MsgStatus {
	switch status {
	case systemv1.SystemMsgStatus_MsgStatus_Revoked:
		return vo.MsgStatusRecalled
	default:
		return vo.MsgStatusNormal
	}
}
