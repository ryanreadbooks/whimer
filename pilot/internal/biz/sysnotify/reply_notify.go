package sysnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/push"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type NotifyUserReplyReq struct {
	Loc            model.NotifyMsgLocation `json:"loc"`
	TargetComment  int64                   `json:"target,omitempty"` // 被回复的评论
	TriggerComment int64                   `json:"trigger"`          // 用这条评论回复的
	SrcUid         int64                   `json:"src_uid"`
	RecvUid        int64                   `json:"recv_uid"`
	NoteId         imodel.NoteId           `json:"note_id"`
	Content        []byte                  `json:"content"` // see model.CommentContent
}

// 通知用户被回复了
func (b *Biz) NotifyUserReply(ctx context.Context, req *NotifyUserReplyReq) error {
	content, err := json.Marshal(req)
	if err != nil {
		return xerror.Wrapf(err, "json marshal reply req failed").WithCtx(ctx)
	}

	reqContent := &systemv1.ReplyMsgContent{
		Uid:       req.SrcUid,
		TargetUid: req.RecvUid,
		Content:   content,
	}

	resp, err := dep.SystemNotifier().NotifyReplyMsg(ctx, &systemv1.NotifyReplyMsgRequest{
		Contents: []*systemv1.ReplyMsgContent{
			reqContent,
		},
	})
	if err != nil {
		return xerror.Wrapf(err, "system notify reply msg failed")
	}

	msgIds := resp.GetMsgIds()[req.RecvUid]
	if len(msgIds.Items) > 0 {
		// 通知用户拉信息
		err = push.PushSysCmdPullUnreadAction(ctx, req.RecvUid)
		if err != nil {
			xlog.Msg("push sys reply notification failed").Extras("recv_uid", req.RecvUid).Errorx(ctx)
			return xerror.Wrapf(err, "push sys reply notification failed").WithCtx(ctx)
		}
	}

	return nil
}
