package whisper

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	whispermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
)

// 发起单聊
func (b *Biz) CreateP2PChat(ctx context.Context, uid, target int64) (string, error) {
	// check target
	_, err := dep.Userer().GetUser(ctx, &userv1.GetUserRequest{Uid: target})
	if err != nil {
		return "", xerror.Wrapf(err, "remote get user failed")
	}

	resp, err := dep.UserChatter().CreateP2PChat(ctx, &userchatv1.CreateP2PChatRequest{
		Uid:    uid,
		Target: target,
	})
	if err != nil {
		return "", xerror.Wrapf(err, "remote create p2p chat failed").WithCtx(ctx)
	}

	return resp.ChatId, nil
}

// 创建群聊
func (b *Biz) CreateGroupChat(ctx context.Context) (string, error) {
	// TODO
	return "", fmt.Errorf("not implemented")
}

// 发消息
func (b *Biz) SendChatMsg(ctx context.Context, chatId, cid string, msg *whispermodel.Msg) (string, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	req := &userchatv1.SendMsgToChatRequest{
		Sender: uid,
		ChatId: chatId,
		Msg: &userchatv1.MsgReq{
			Cid:  cid,
			Type: pbmsg.MsgType(msg.Type),
			// Content: userchatv1.isMsgReq_Content, // content is assigned below
		},
	}
	err := whispermodel.AssignPbMsgContent(msg, req.Msg)
	if err != nil {
		return "", err
	}

	resp, err := dep.UserChatter().SendMsgToChat(ctx, req)
	if err != nil {
		return "", xerror.Wrapf(err, "remove send msg to chat failed").WithCtx(ctx)
	}

	membersResp, err := dep.UserChatter().GetChatMembers(ctx,
		&userchatv1.GetChatMembersRequest{
			ChatId: chatId,
		})
	if err != nil {
		xlog.Msg("remote get chat members failed").Extras("chat_id", chatId).Errorx(ctx)
	} else {
		members := xslice.Filter(membersResp.GetMembers(), func(_ int, v int64) bool { return v == uid })
		if len(members) == 0 {
			xlog.Msgf("chat members of chat %s is empty", chatId).Errorx(ctx)
		} else {
			// 消息推送
			if err := pushcenter.BatchNotifyWhisperMsg(ctx, members); err != nil {
				xlog.Msg("push notify whisper msg failed").Extras("chat_id", chatId, "members", members).Errorx(ctx)
			}
		}
	}

	return resp.MsgId, nil
}
