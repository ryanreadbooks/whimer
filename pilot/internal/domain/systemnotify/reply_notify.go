package systemnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// 通知用户被回复了
func (b *DomainService) NotifyUserReply(ctx context.Context, req *vo.NotifyUserReplyParam) error {
	content, err := json.Marshal(req)
	if err != nil {
		return xerror.Wrapf(err, "json marshal reply req failed").WithCtx(ctx)
	}

	msgId, err := b.systemNotifyAdapter.NotifyReplyMsg(ctx, &vo.SystemMessage{
		Uid:       req.SrcUid,
		TargetUid: req.RecvUid,
		Content:   content,
	})
	if err != nil {
		return xerror.Wrapf(err, "system notify reply msg failed")
	}

	if msgId != "" {
		// 通知用户拉信息
		err = pushcenter.NotifySystemMsg(ctx, req.RecvUid)
		if err != nil {
			xlog.Msg("push sys reply notification failed").Extras("recv_uid", req.RecvUid).Errorx(ctx)
			return xerror.Wrapf(err, "push sys reply notification failed").WithCtx(ctx)
		}
	}

	return nil
}
