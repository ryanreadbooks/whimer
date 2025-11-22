package sysnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type NotifyLikesOnNoteReq struct {
	NoteId imodel.NoteId `json:"note_id"`
}

type NotifyLikesOnCommentReq struct {
	NoteId    imodel.NoteId `json:"note_id"`
	CommentId int64         `json:"comment_id"`
}

type notifyLikesContent struct {
	*NotifyLikesOnNoteReq    `json:"note_content,omitempty"`
	*NotifyLikesOnCommentReq `json:"comment_content,omitempty"`

	Loc     model.NotifyMsgLocation `json:"loc"`
	Uid     int64                   `json:"uid"`      // 谁点赞
	RecvUid int64                   `json:"recv_uid"` // 谁被点赞
}

// 通知用户笔记收到点赞
func (b *Biz) NotifyUserLikesOnNote(ctx context.Context, uid, recvUid int64, req *NotifyLikesOnNoteReq) error {
	lReq := &notifyLikesContent{
		NotifyLikesOnNoteReq: req,
		Loc:                  model.NotifyMsgOnNote,
		Uid:                  uid,
		RecvUid:              recvUid,
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
func (b *Biz) NotifyUserLikesOnComment(ctx context.Context, uid, recvUid int64, req *NotifyLikesOnCommentReq) error {
	lReq := &notifyLikesContent{
		NotifyLikesOnCommentReq: req,
		Loc:                     model.NotifyMsgOnComment,
		Uid:                     uid,
		RecvUid:                 recvUid,
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

func (b *Biz) notifyLikesAndPush(ctx context.Context, uid, recvUid int64, content []byte) error {
	resp, err := dep.SystemNotifier().NotifyLikesMsg(ctx, &systemv1.NotifyLikesMsgRequest{
		Contents: []*systemv1.LikeMsgContent{
			{
				Uid:       uid,
				TargetUid: recvUid,
				Content:   content,
			},
		},
	})
	if err != nil {
		return xerror.Wrapf(err, "sys notify likes msg failed").WithCtx(ctx)
	}

	// 通知用户拉消息
	if msgIds, ok := resp.GetMsgIds()[recvUid]; ok && len(msgIds.GetItems()) > 0 {
		err = pushcenter.NotifySystemMsg(ctx, recvUid)
		if err != nil {
			return xerror.Wrapf(err, "push sys likes notification failed").
				WithExtras("recv_uid", recvUid).WithCtx(ctx)
		}
	}

	return nil
}
