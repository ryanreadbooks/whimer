package systemnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type notifyLikesContent struct {
	*vo.NotifyLikesOnNoteParam    `json:"note_content,omitempty"`
	*vo.NotifyLikesOnCommentParam `json:"comment_content,omitempty"`

	Loc     model.NotifyMsgLocation `json:"loc"`
	Uid     int64                   `json:"uid"`      // 谁点赞
	RecvUid int64                   `json:"recv_uid"` // 谁被点赞
}

// 通知用户笔记收到点赞
func (b *DomainService) NotifyUserLikesOnNote(ctx context.Context, uid, recvUid int64, req *vo.NotifyLikesOnNoteParam) error {
	lReq := &notifyLikesContent{
		NotifyLikesOnNoteParam: req,
		Loc:                    model.NotifyMsgOnNote,
		Uid:                    uid,
		RecvUid:                recvUid,
	}

	content, err := json.Marshal(lReq)
	if err != nil {
		return xerror.Wrapf(err, "json marshal notify likes content on note failed").WithCtx(ctx)
	}

	if err := b.notifyLikesAndPush(ctx, uid, recvUid, content); err != nil {
		return err
	}

	return nil
}

// 通知用户评论收到点赞
func (b *DomainService) NotifyUserLikesOnComment(ctx context.Context, uid, recvUid int64, req *vo.NotifyLikesOnCommentParam) error {
	lReq := &notifyLikesContent{
		NotifyLikesOnCommentParam: req,
		Loc:                       model.NotifyMsgOnComment,
		Uid:                       uid,
		RecvUid:                   recvUid,
	}

	content, err := json.Marshal(lReq)
	if err != nil {
		return xerror.Wrapf(err, "json marshal notify likes content on comment failed").WithCtx(ctx)
	}

	if err := b.notifyLikesAndPush(ctx, uid, recvUid, content); err != nil {
		return err
	}

	return nil
}

func (b *DomainService) notifyLikesAndPush(ctx context.Context, uid, recvUid int64, content []byte) error {
	msgId, err := b.systemNotifyAdapter.NotifyLikesMsg(ctx, &vo.SystemMessage{
		Uid:       uid,
		TargetUid: recvUid,
		Content:   content,
	})
	if err != nil {
		return xerror.Wrapf(err, "sys notify likes msg failed").WithCtx(ctx)
	}

	// 通知用户拉消息
	if msgId != "" {
		err = pushcenter.NotifySystemMsg(ctx, recvUid)
		if err != nil {
			return xerror.Wrapf(err, "push sys likes notification failed").
				WithExtras("recv_uid", recvUid).WithCtx(ctx)
		}
	}

	return nil
}
