package whisper

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	whispermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/whisper/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	globalmodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
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
func (b *Biz) SendChatMsg(ctx context.Context, chatId, cid string, msg *whispermodel.MsgReq) (string, error) {
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
	err := whispermodel.AssignPbMsgReqContent(msg, req.Msg)
	if err != nil {
		return "", err
	}

	resp, err := dep.UserChatter().SendMsgToChat(ctx, req)
	if err != nil {
		return "", xerror.Wrapf(err, "remove send msg to chat failed").WithCtx(ctx)
	}

	b.asyncNotifyWhisperEvent(ctx, uid, chatId)
	return resp.MsgId, nil
}

func (b *Biz) asyncNotifyWhisperEvent(ctx context.Context, uid int64, chatId string) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "whisper.biz.notify_whisper",
		Job: func(ctx context.Context) error {
			req := &userchatv1.GetChatMembersRequest{
				ChatId: chatId,
			}
			resp, err := dep.UserChatter().GetChatMembers(ctx, req)
			if err != nil {
				xlog.Msg("remote get chat members failed").Extras("chat_id", chatId).Errorx(ctx)
			} else {
				// 排除uid 一般是消息发送者或者消息撤回者
				members := xslice.Filter(resp.GetMembers(), func(_ int, v int64) bool { return v == uid })
				// 消息推送
				if err := pushcenter.BatchNotifyWhisperMsg(ctx, members); err != nil {
					xlog.Msg("push notify whisper msg failed").Extras("chat_id", chatId, "members", members).Errorx(ctx)
				}
			}

			return nil
		},
	})
}

// 列出用户最近会话列表
func (b *Biz) ListRecentChats(ctx context.Context, uid int64,
	cursor string, cnt int32) ([]*whispermodel.RecentChat, *globalmodel.ListResult[string], error) {
	resp, err := dep.UserChatter().ListRecentChats(ctx,
		&userchatv1.ListRecentChatsRequest{
			Uid:    uid,
			Cursor: cursor,
			Count:  cnt,
		})
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "remote list recent chats failed").WithCtx(ctx)
	}

	p2pChatIds := make([]string, 0, len(resp.RecentChats))
	for _, r := range resp.RecentChats {
		if r.ChatType == userchatv1.ChatType_P2P {
			p2pChatIds = append(p2pChatIds, r.ChatId)
		}
	}

	p2pMembersResp, err := dep.UserChatter().BatchGetChatMembers(ctx,
		&userchatv1.BatchGetChatMembersRequest{
			ChatIds: p2pChatIds,
		})
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "remote batch get members failed").WithCtx(ctx)
	}

	peers := make(map[string]int64, len(p2pMembersResp.GetMembersMap()))
	for _, p2pChatId := range p2pChatIds {
		mis, ok := p2pMembersResp.GetMembersMap()[p2pChatId]
		if ok {
			memberIds := mis.GetInts()
			// 单聊的另外一个人
			peerMembers := xslice.Filter(memberIds, func(_ int, v int64) bool { return v == uid })
			if len(peerMembers) > 0 {
				peers[p2pChatId] = peerMembers[0]
			}
		}
	}

	userInfos, err := dep.Userer().BatchGetUserV2(ctx,
		&userv1.BatchGetUserV2Request{
			Uids: xmap.Values(peers),
		})
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "remote batch get user failed").WithCtx(ctx)
	}
	chatPeerUsers := make(map[string]*userv1.UserInfo)
	for chatId, userId := range peers {
		chatPeerUsers[chatId] = userInfos.GetUsers()[userId]
	}

	// format recent chats
	recentChats := make([]*whispermodel.RecentChat, 0, len(resp.RecentChats))
	for _, pbRecentChat := range resp.RecentChats {
		recentChat := &whispermodel.RecentChat{
			Uid:         pbRecentChat.Uid,
			ChatId:      pbRecentChat.ChatId,
			ChatType:    whispermodel.ChatTypeFromPb(pbRecentChat.ChatType),
			ChatName:    pbRecentChat.ChatName,
			ChatCreator: pbRecentChat.ChatCreator,
			UnreadCount: pbRecentChat.UnreadCount,
			Mtime:       pbRecentChat.Mtime,
			IsPinned:    pbRecentChat.IsPinned,
			LastMsg:     whispermodel.MsgFromChatMsgPb(pbRecentChat.GetLastMsg()),
		}
		// 替换p2p chat的name
		if recentChat.ChatType == whispermodel.P2PChat {
			peer := chatPeerUsers[pbRecentChat.ChatId]
			recentChat.ChatName = peer.GetNickname()
			recentChat.Cover = peer.GetAvatar()
			recentChat.Peer = peer
		}
		recentChats = append(recentChats, recentChat)
	}

	return recentChats, &globalmodel.ListResult[string]{
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (b *Biz) ListChatMsgs(ctx context.Context, uid int64, chatId string,
	pos int64, cnt int32) ([]*whispermodel.Msg, error) {

	resp, err := dep.UserChatter().ListChatMsgs(ctx, &userchatv1.ListChatMsgsRequest{
		ChatId: chatId,
		Uid:    uid,
		Pos:    pos,
		Count:  cnt,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote list chat msgs failed").WithCtx(ctx)
	}

	msgs := make([]*whispermodel.Msg, 0, len(resp.GetChatMsgs()))
	for _, m := range resp.GetChatMsgs() {
		msgs = append(msgs, whispermodel.MsgFromChatMsgPb(m))
	}

	return msgs, nil
}

// 撤回消息
func (b *Biz) RecallChatMsg(ctx context.Context, chatId, msgId string) error {
	var (
		uid = metadata.Uid(ctx)
	)

	_, err := dep.UserChatter().RecallMsg(ctx, &userchatv1.RecallMsgRequest{
		Uid:    uid,
		MsgId:  msgId,
		ChatId: chatId,
	})
	if err != nil {
		return xerror.Wrapf(err, "remote recall msg failed").
			WithCtx(ctx).WithExtras("msg_id", msgId, "chat_id", chatId)
	}

	// 消息推送
	b.asyncNotifyWhisperEvent(ctx, uid, chatId)

	return nil
}

// 清除用户会话未读数
func (b *Biz) ClearChatUnread(ctx context.Context, chatId string) error {
	var (
		uid = metadata.Uid(ctx)
	)

	_, err := dep.UserChatter().ClearChatUnread(ctx, &userchatv1.ClearChatUnreadRequest{
		ChatId: chatId,
		Uid:    uid,
	})

	if err != nil {
		return xerror.Wrapf(err, "remote clear chat unread failed").
			WithCtx(ctx).WithExtras("chat_id", chatId)
	}

	return nil
}
