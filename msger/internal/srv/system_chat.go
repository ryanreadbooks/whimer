package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
)

type SystemChatSrv struct {
	chatBiz bizsyschat.ChatBiz
}

func NewSystemChatSrv(biz biz.Biz) *SystemChatSrv {
	return &SystemChatSrv{
		chatBiz: biz.SystemBiz,
	}
}

func (s *SystemChatSrv) SendMsg(ctx context.Context, req bizsyschat.CreateSystemMsgReq) (uuid.UUID, error) {
	// TODO uid 系统通知设置检查
	msgId, err := s.chatBiz.CreateMsg(ctx, &req)
	if err != nil {
		return uuid.UUID{}, xerror.Wrapf(err, "srv failed to send system msg").WithCtx(ctx)
	}

	// TODO 消息推送

	return msgId.Id, nil
}

// 发送通用系统消息
func (s *SystemChatSrv) NotifyCommonSystemMsg(ctx context.Context, uid int64) error {

	return nil
}

// 发送回复我的系统消息
func (s *SystemChatSrv) NotifyReplyToMeSystemMsg(ctx context.Context, uid int64) error {

	return nil
}

// 发送@我的系统消息
func (s *SystemChatSrv) NotifyMentionedByOthersSystemMsg(ctx context.Context, uid int64) error {

	return nil
}

// 发送收到的赞系统消息
func (s *SystemChatSrv) NotifyLikeReceivedSystemMsg(ctx context.Context, uid int64) error {

	return nil
}
