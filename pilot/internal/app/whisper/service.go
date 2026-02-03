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
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/whisper/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/pushcenter"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	uservo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

type Service struct {
	whisperAdapter repository.UserChatAdapter
	userAdapter    userrepo.UserServiceAdapter
}

func NewService(
	whisperAdapter repository.UserChatAdapter,
	userAdapter userrepo.UserServiceAdapter,
) *Service {
	return &Service{
		whisperAdapter: whisperAdapter,
		userAdapter:    userAdapter,
	}
}

func (s *Service) CreateP2PChat(ctx context.Context, cmd *dto.CreateP2PChatCommand) (string, error) {
	if _, err := s.userAdapter.GetUser(ctx, cmd.Target); err != nil {
		return "", xerror.Wrapf(err, "get target user failed")
	}
	return s.whisperAdapter.CreateP2PChat(ctx, cmd.Uid, cmd.Target)
}

func (s *Service) CreateGroupChat(ctx context.Context) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *Service) SendChatMsg(ctx context.Context, cmd *dto.SendChatMsgCommand) (string, error) {
	uid := metadata.Uid(ctx)

	msgReq := cmd.ToMsgReq()
	if err := msgReq.Validate(ctx); err != nil {
		return "", err
	}

	msgId, err := s.whisperAdapter.SendMsgToChat(ctx, &repository.SendMsgParams{
		Sender:  uid,
		ChatId:  cmd.ChatId,
		Cid:     msgReq.Cid,
		Type:    msgReq.Type,
		Content: msgReq.Content,
	})
	if err != nil {
		return "", xerror.Wrapf(err, "send msg to chat failed").WithCtx(ctx)
	}

	s.asyncNotifyWhisperEvent(ctx, uid, cmd.ChatId)
	return msgId, nil
}

func (s *Service) asyncNotifyWhisperEvent(ctx context.Context, uid int64, chatId string) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "whisper.app.notify_whisper",
		Job: func(ctx context.Context) error {
			members, err := s.whisperAdapter.GetChatMembers(ctx, chatId)
			if err != nil {
				xlog.Msg("get chat members failed").Extras("chat_id", chatId).Errorx(ctx)
			} else {
				members = xslice.Filter(members, func(_ int, v int64) bool { return v == uid })
				if err := pushcenter.BatchNotifyWhisperMsg(ctx, members); err != nil {
					xlog.Msg("push notify whisper msg failed").Extras("chat_id", chatId, "members", members).Errorx(ctx)
				}
			}
			return nil
		},
	})
}

func (s *Service) ListRecentChats(ctx context.Context, query *dto.ListRecentChatsQuery) (*dto.ListRecentChatsResult, error) {
	resp, err := s.whisperAdapter.ListRecentChats(ctx, query.Uid, query.Cursor, query.Count)
	if err != nil {
		return nil, xerror.Wrapf(err, "list recent chats failed").WithCtx(ctx)
	}

	// 收集 p2p 聊天的 peer uids
	p2pChatIds := make([]string, 0)
	for _, chat := range resp.Chats {
		if chat.ChatType == vo.P2PChat {
			p2pChatIds = append(p2pChatIds, chat.ChatId)
		}
	}

	// 获取 p2p 聊天成员
	peers := make(map[string]int64)
	if len(p2pChatIds) > 0 {
		membersMap, err := s.whisperAdapter.BatchGetChatMembers(ctx, p2pChatIds)
		if err != nil {
			return nil, xerror.Wrapf(err, "batch get chat members failed").WithCtx(ctx)
		}
		for chatId, memberIds := range membersMap {
			peerMembers := xslice.Filter(memberIds, func(_ int, v int64) bool { return v == query.Uid })
			if len(peerMembers) > 0 {
				peers[chatId] = peerMembers[0]
			}
		}
	}

	// 获取用户信息
	peerUids := xmap.Values(peers)
	userInfos, err := s.userAdapter.BatchGetUser(ctx, peerUids)
	if err != nil {
		return nil, xerror.Wrapf(err, "batch get user failed").WithCtx(ctx)
	}

	// 组装结果
	chats := make([]*dto.RecentChatWithPeer, 0, len(resp.Chats))
	for _, chat := range resp.Chats {
		result := &dto.RecentChatWithPeer{
			ChatId:      chat.ChatId,
			ChatType:    chat.ChatType,
			ChatName:    chat.ChatName,
			ChatCreator: chat.ChatCreator,
			UnreadCount: chat.UnreadCount,
			Mtime:       chat.Mtime,
			IsPinned:    chat.IsPinned,
			Cover:       chat.Cover,
			LastMsg:     dto.MsgWithSenderFromEntity(chat.LastMsg),
		}

		if chat.ChatType == vo.P2PChat {
			if peerUid, ok := peers[chat.ChatId]; ok {
				if peer, ok := userInfos[peerUid]; ok {
					result.ChatName = peer.Nickname
					result.Cover = peer.Avatar
					result.Peer = &commondto.User{
						Uid:       peer.Uid,
						Nickname:  peer.Nickname,
						Avatar:    peer.Avatar,
						StyleSign: peer.StyleSign,
					}
				}
			}
		}
		chats = append(chats, result)
	}

	return &dto.ListRecentChatsResult{
		Chats:      chats,
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (s *Service) ListChatMsgs(ctx context.Context, query *dto.ListChatMsgsQuery) ([]*dto.MsgWithSender, error) {
	msgs, err := s.whisperAdapter.ListChatMsgs(ctx, query.ChatId, query.Uid, query.Pos, query.Count)
	if err != nil {
		return nil, xerror.Wrapf(err, "list chat msgs failed").WithCtx(ctx)
	}

	// 收集 sender uids 并获取用户信息
	senderUids := xslice.Uniq(xslice.Extract(msgs, func(m *entity.Msg) int64 { return m.SenderUid }))
	userInfos, err := s.userAdapter.BatchGetUser(ctx, senderUids)
	if err != nil {
		xlog.Msg("batch get user failed").Err(err).Errorx(ctx)
		userInfos = make(map[int64]*uservo.User)
	}

	// 组装结果
	result := make([]*dto.MsgWithSender, 0, len(msgs))
	for _, msg := range msgs {
		item := dto.MsgWithSenderFromEntity(msg)
		if sender, ok := userInfos[msg.SenderUid]; ok {
			item.Sender = sender
		}
		result = append(result, item)
	}

	return result, nil
}

func (s *Service) RecallChatMsg(ctx context.Context, cmd *dto.RecallChatMsgCommand) error {
	uid := metadata.Uid(ctx)
	if err := s.whisperAdapter.RecallMsg(ctx, uid, cmd.ChatId, cmd.MsgId); err != nil {
		return xerror.Wrapf(err, "recall msg failed").WithCtx(ctx).WithExtras("msg_id", cmd.MsgId, "chat_id", cmd.ChatId)
	}
	s.asyncNotifyWhisperEvent(ctx, uid, cmd.ChatId)
	return nil
}

func (s *Service) ClearChatUnread(ctx context.Context, cmd *dto.ClearChatUnreadCommand) error {
	uid := metadata.Uid(ctx)
	return s.whisperAdapter.ClearChatUnread(ctx, uid, cmd.ChatId)
}
