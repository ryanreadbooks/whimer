package systemnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/systemnotify/dto"
	commentrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	noterepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/repository"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify"
	notifyentity "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
	userrepo "github.com/ryanreadbooks/whimer/pilot/internal/domain/user/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dao/kafka"
	sysmsgkfkdao "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dao/kafka/sysmsg"
	"golang.org/x/sync/errgroup"
)

type Service struct {
	domainService    *systemnotify.DomainService
	noteFeedAdapter  noterepo.NoteFeedAdapter
	noteLikesAdapter noterepo.NoteLikesAdapter
	commentAdapter   commentrepo.CommentAdapter
	userAdapter      userrepo.UserServiceAdapter
}

func NewService(
	domainService *systemnotify.DomainService,
	noteFeedAdapter noterepo.NoteFeedAdapter,
	noteLikesAdapter noterepo.NoteLikesAdapter,
	commentAdapter commentrepo.CommentAdapter,
	userAdapter userrepo.UserServiceAdapter,
) *Service {
	return &Service{
		domainService:    domainService,
		noteFeedAdapter:  noteFeedAdapter,
		noteLikesAdapter: noteLikesAdapter,
		commentAdapter:   commentAdapter,
		userAdapter:      userAdapter,
	}
}

// 获取用户的被@消息列表
func (s *Service) ListUserMentionMsg(ctx context.Context,
	uid int64, cursor string, count int32,
) (*dto.ListUserMentionMsgResult, error) {
	resp, err := s.domainService.ListMentionMsg(ctx, uid, cursor, count)
	if err != nil {
		return nil, xerror.Wrapf(err, "list mention msg failed").
			WithExtras("uid", uid, "cursor", cursor, "count", count).WithCtx(ctx)
	}

	msgs := parseMentionMsgs(ctx, resp.Messages)

	// 懒加载检查
	if err := s.lazyCheckMentionMsgSource(ctx, metadata.Uid(ctx), msgs); err != nil {
		return nil, xerror.Wrapf(err, "lazy check mention source failed")
	}

	// 附加用户信息
	msgsWithUser, err := s.attachUserToMentionMsgs(ctx, msgs)
	if err != nil {
		return nil, xerror.Wrapf(err, "attach user to mention msgs failed").WithCtx(ctx)
	}

	return &dto.ListUserMentionMsgResult{
		Msgs:    msgsWithUser,
		ChatId:  resp.ChatId,
		HasNext: resp.HasMore,
	}, nil
}

// 获取用户被回复的消息列表
func (s *Service) ListUserReplyMsg(ctx context.Context,
	uid int64, cursor string, count int32,
) (*dto.ListUserReplyMsgResult, error) {
	resp, err := s.domainService.ListReplyMsg(ctx, uid, cursor, count)
	if err != nil {
		return nil, xerror.Wrapf(err, "list reply msg failed").
			WithExtras("uid", uid, "cursor", cursor, "count", count).WithCtx(ctx)
	}

	msgs := parseReplyMsgs(ctx, uid, resp.Messages)

	if err := s.lazyCheckReplyMsgSource(ctx, metadata.Uid(ctx), msgs); err != nil {
		return nil, xerror.Wrapf(err, "lazy check reply source failed").WithCtx(ctx)
	}

	// 附加用户信息
	msgsWithUser, err := s.attachUserToReplyMsgs(ctx, msgs)
	if err != nil {
		return nil, xerror.Wrapf(err, "attach user to reply msgs failed").WithCtx(ctx)
	}

	return &dto.ListUserReplyMsgResult{
		Msgs:    msgsWithUser,
		ChatId:  resp.ChatId,
		HasNext: resp.HasMore,
	}, nil
}

// 获取用户收到的赞消息列表
func (s *Service) ListUserLikeMsg(ctx context.Context, uid int64, cursor string) (*dto.ListUserLikesMsgResult, error) {
	resp, err := s.domainService.ListLikesMsg(ctx, uid, cursor, 20)
	if err != nil {
		return nil, xerror.Wrapf(err, "list likes msg failed").
			WithExtras("uid", uid, "cursor", cursor).WithCtx(ctx)
	}

	msgs, noteLikings, commentLikings := parseLikesMsgs(ctx, resp.Messages)

	if err := s.lazyCheckLikesMsgSource(ctx, metadata.Uid(ctx), msgs, noteLikings, commentLikings); err != nil {
		return nil, xerror.Wrapf(err, "lazy check likes source failed").WithCtx(ctx)
	}

	// 附加用户信息
	msgsWithUser, err := s.attachUserToLikesMsgs(ctx, msgs)
	if err != nil {
		return nil, xerror.Wrapf(err, "attach user to likes msgs failed").WithCtx(ctx)
	}

	return &dto.ListUserLikesMsgResult{
		Msgs:    msgsWithUser,
		ChatId:  resp.ChatId,
		HasNext: resp.HasMore,
	}, nil
}

func (s *Service) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	return s.domainService.ClearChatUnread(ctx, uid, chatId)
}

func (s *Service) GetChatUnread(ctx context.Context, uid int64) (*notifyentity.ChatsUnreadCount, error) {
	return s.domainService.GetChatUnread(ctx, uid)
}

func (s *Service) DeleteSysMsg(ctx context.Context, uid int64, msgId string) error {
	return s.domainService.DeleteSysMsg(ctx, uid, msgId)
}

func (s *Service) lazyCheckMentionMsgSource(ctx context.Context, uid int64, msgs []*notifyentity.MentionedMsg) error {
	noteIds, commentIds := systemnotify.ExtractMentionMsgSourceIds(msgs)
	noteIds = xslice.FilterZero(xslice.Uniq(noteIds))
	commentIds = xslice.FilterZero(xslice.Uniq(commentIds))

	noteExistence, commentExistence, err := s.checkSourcesExistence(ctx, noteIds, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "check sources existence failed").WithCtx(ctx)
	}

	pendingMsgIds := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if msg.ShouldRuleOut(noteExistence, commentExistence) {
			pendingMsgIds = append(pendingMsgIds, msg.Id)
		}
	}

	s.asyncBatchDeleteMsgs(ctx, uid, pendingMsgIds)
	return nil
}

func (s *Service) lazyCheckReplyMsgSource(ctx context.Context, uid int64, msgs []*notifyentity.ReplyMsg) error {
	noteIds, commentIds := systemnotify.ExtractReplyMsgSourceIds(msgs)
	noteIds = xslice.FilterZero(xslice.Uniq(noteIds))
	commentIds = xslice.FilterZero(xslice.Uniq(commentIds))

	noteExistence, commentExistence, err := s.checkSourcesExistence(ctx, noteIds, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "check sources existence failed").WithCtx(ctx)
	}

	pendingMsgIds := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if msg.ShouldRuleOut(noteExistence, commentExistence) {
			pendingMsgIds = append(pendingMsgIds, msg.Id)
		}
	}

	s.asyncBatchDeleteMsgs(ctx, uid, pendingMsgIds)
	return nil
}

func (s *Service) lazyCheckLikesMsgSource(ctx context.Context, uid int64, msgs []*notifyentity.LikesMsg,
	noteLikings, commentLikings map[int64][]int64,
) error {
	noteLikeStatus, commentLikeStatus, err := s.checkLikeStatus(ctx, noteLikings, commentLikings)
	if err != nil {
		return xerror.Wrapf(err, "check like status failed").WithCtx(ctx)
	}

	pendingMsgIds := make([]string, 0, len(msgs))
	for _, msg := range msgs {
		if msg.ShouldFilterByLikeStatus(noteLikeStatus, commentLikeStatus) {
			pendingMsgIds = append(pendingMsgIds, msg.Id)
		}
	}

	noteIds, commentIds := systemnotify.ExtractLikesMsgSourceIds(msgs)
	noteIds = xslice.FilterZero(xslice.Uniq(noteIds))
	commentIds = xslice.FilterZero(xslice.Uniq(commentIds))

	noteExistence, commentExistence, err := s.checkSourcesExistence(ctx, noteIds, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "check sources existence failed").WithCtx(ctx)
	}

	for _, msg := range msgs {
		if msg.ShouldRuleOut(noteExistence, commentExistence) {
			pendingMsgIds = append(pendingMsgIds, msg.Id)
		}
	}

	s.asyncBatchDeleteMsgs(ctx, uid, pendingMsgIds)
	return nil
}

func (s *Service) checkSourcesExistence(ctx context.Context, noteIds, commentIds []int64) (
	noteExistence, commentExistence map[int64]bool, err error,
) {
	noteExistence = make(map[int64]bool)
	commentExistence = make(map[int64]bool)

	eg, ctx := errgroup.WithContext(ctx)

	if len(noteIds) > 0 {
		eg.Go(recovery.DoV2(func() error {
			result, err := s.noteFeedAdapter.BatchCheckNoteExist(ctx, noteIds)
			if err != nil {
				return xerror.Wrapf(err, "batch check note exist failed")
			}
			noteExistence = result
			return nil
		}))
	}

	if len(commentIds) > 0 {
		eg.Go(recovery.DoV2(func() error {
			result, err := s.commentAdapter.BatchCheckCommentExist(ctx, commentIds)
			if err != nil {
				return xerror.Wrapf(err, "batch check comment exist failed")
			}
			commentExistence = result
			return nil
		}))
	}

	if err := eg.Wait(); err != nil {
		return nil, nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return noteExistence, commentExistence, nil
}

func (s *Service) checkLikeStatus(ctx context.Context,
	noteLikings, commentLikings map[int64][]int64,
) (noteLikeStatus, commentLikeStatus map[int64]map[int64]bool, err error) {
	noteLikeStatus = make(map[int64]map[int64]bool)
	commentLikeStatus = make(map[int64]map[int64]bool)

	eg, ctx := errgroup.WithContext(ctx)

	if len(noteLikings) > 0 {
		eg.Go(recovery.DoV2(func() error {
			result, err := s.noteLikesAdapter.BatchCheckUserLikeStatus(ctx, noteLikings)
			if err != nil {
				return xerror.Wrapf(err, "batch check note like status failed")
			}
			noteLikeStatus = result
			return nil
		}))
	}

	if len(commentLikings) > 0 {
		eg.Go(recovery.DoV2(func() error {
			result, err := s.commentAdapter.BatchCheckUsersLikeComment(ctx, commentLikings)
			if err != nil {
				return xerror.Wrapf(err, "batch check comment like status failed")
			}
			commentLikeStatus = result
			return nil
		}))
	}

	if err := eg.Wait(); err != nil {
		return nil, nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return noteLikeStatus, commentLikeStatus, nil
}

func (s *Service) asyncBatchDeleteMsgs(ctx context.Context, uid int64, msgIds []string) {
	if len(msgIds) == 0 {
		return
	}

	xlog.Msgf("sysmsg pending delete msgids length = %d", len(msgIds)).Debugx(ctx)

	deletions := make([]*sysmsgkfkdao.DeletionEvent, 0, len(msgIds))
	for _, msgId := range msgIds {
		if uid != 0 && msgId != "" {
			deletions = append(deletions, &sysmsgkfkdao.DeletionEvent{
				Uid:   uid,
				MsgId: msgId,
			})
		}
	}

	if err := kafka.Dao().SysMsgEventProducer.AsyncPutDeletion(ctx, deletions); err != nil {
		xlog.Msg("sysmsg async put deletion failed").Err(err).Extras("args", deletions).Errorx(ctx)
	}
}

type mentionMsgContent struct {
	*notifyvo.NotifyAtUsersOnNoteParamContent    `json:"note_content,omitempty"`
	*notifyvo.NotifyAtUsersOnCommentParamContent `json:"comment_content,omitempty"`
	Receivers                                    []*mentionvo.AtUser        `json:"receivers"`
	Loc                                          notifyvo.NotifyMsgLocation `json:"loc"`
}

func parseMentionMsgs(ctx context.Context, rawMsgs []*notifyvo.RawSystemMsg) []*notifyentity.MentionedMsg {
	msgs := make([]*notifyentity.MentionedMsg, 0, len(rawMsgs))

	for _, msg := range rawMsgs {
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			xlog.Msg("parse mention msg id failed").Err(err).Extras("msgid", msg.Id).Errorx(ctx)
			continue
		}

		mm := notifyentity.MentionedMsg{
			Id:     msg.Id,
			SendAt: mgid.UnixSec(),
		}

		if msg.Status != notifyvo.MsgStatusRecalled {
			var v mentionMsgContent
			if err := json.Unmarshal(msg.Content, &v); err != nil {
				xlog.Msg("unmarshal mention msg content failed").Err(err).Errorx(ctx)
				continue
			}

			var (
				loc       notifyvo.NotifyMsgLocation
				sourceUid int64
				noteId    notevo.NoteId
				content   string
				commentId int64
			)

			if v.NotifyAtUsersOnNoteParamContent != nil {
				loc = notifyvo.NotifyMsgOnNote
				sourceUid = v.NotifyAtUsersOnNoteParamContent.SourceUid
				noteId = v.NotifyAtUsersOnNoteParamContent.NoteId
				content = v.NotifyAtUsersOnNoteParamContent.NoteDesc
			} else if v.NotifyAtUsersOnCommentParamContent != nil {
				loc = notifyvo.NotifyMsgOnComment
				sourceUid = v.NotifyAtUsersOnCommentParamContent.SourceUid
				content = v.NotifyAtUsersOnCommentParamContent.Comment
				commentId = v.NotifyAtUsersOnCommentParamContent.CommentId
				noteId = v.NotifyAtUsersOnCommentParamContent.NoteId
			}

			mm.Type = loc
			mm.Uid = sourceUid
			mm.RecvUsers = v.Receivers
			mm.NoteId = noteId
			mm.CommentId = commentId
			mm.Content = content
			mm.Status = notifyvo.MsgStatusNormal
		} else {
			mm.Status = notifyvo.MsgStatusRecalled
		}
		msgs = append(msgs, &mm)
	}

	return msgs
}

func parseReplyMsgs(ctx context.Context, uid int64, rawMsgs []*notifyvo.RawSystemMsg) []*notifyentity.ReplyMsg {
	msgs := make([]*notifyentity.ReplyMsg, 0, len(rawMsgs))

	for _, msg := range rawMsgs {
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			xlog.Msg("parse reply msg id failed").Err(err).Extras("msgid", msg.Id).Errorx(ctx)
			continue
		}

		rm := notifyentity.ReplyMsg{
			Id:      msg.Id,
			SendAt:  mgid.UnixSec(),
			RecvUid: uid,
		}

		if msg.Status != notifyvo.MsgStatusRecalled {
			var content notifyvo.NotifyUserReplyParam
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				xlog.Msg("unmarshal reply content failed").Extras("msg_id", msg.Id).Errorx(ctx)
				continue
			}

			var commentContent notifyvo.CommentContent
			if err := json.Unmarshal(content.Content, &commentContent); err != nil {
				xlog.Msg("unmarshal comment content failed").Extras("msg_id", msg.Id).Errorx(ctx)
				continue
			}

			rm.NoteId = content.NoteId
			rm.Uid = content.SrcUid
			rm.Status = notifyvo.MsgStatusNormal
			rm.Content = commentContent.Text
			rm.Type = content.Loc
			rm.TargetComment = content.TargetComment
			rm.TriggerComment = content.TriggerComment
			if len(commentContent.AtUsers) > 0 {
				rm.Ext = &notifyentity.ReplyMsgExt{AtUsers: commentContent.AtUsers}
			}
		} else {
			rm.Status = notifyvo.MsgStatusRecalled
		}

		msgs = append(msgs, &rm)
	}

	return msgs
}

func parseLikesMsgs(ctx context.Context, rawMsgs []*notifyvo.RawSystemMsg) (
	[]*notifyentity.LikesMsg, map[int64][]int64, map[int64][]int64,
) {
	msgs := make([]*notifyentity.LikesMsg, 0, len(rawMsgs))
	noteLikings := make(map[int64][]int64, len(rawMsgs))
	commentLikings := make(map[int64][]int64, len(rawMsgs))

	for _, msg := range rawMsgs {
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			xlog.Msg("parse likes msg id failed").Err(err).Extras("msgid", msg.Id).Errorx(ctx)
			continue
		}

		lm := notifyentity.LikesMsg{
			Id:      msg.Id,
			SendAt:  mgid.UnixSec(),
			RecvUid: msg.RecvUid,
		}

		if msg.Status != notifyvo.MsgStatusRecalled {
			var content notifyvo.LikesMessage
			if err := json.Unmarshal(msg.Content, &content); err != nil {
				xlog.Msg("unmarshal likes content failed").Err(err).Errorx(ctx)
				continue
			}

			switch content.Loc {
			case notifyvo.NotifyMsgOnNote:
				lm.NoteId = content.NotifyLikesOnNoteParam.NoteId
				noteLikings[content.Uid] = append(noteLikings[content.Uid], int64(lm.NoteId))
			case notifyvo.NotifyMsgOnComment:
				lm.NoteId = content.NotifyLikesOnCommentParam.NoteId
				lm.CommentId = content.NotifyLikesOnCommentParam.CommentId
				commentLikings[content.Uid] = append(commentLikings[content.Uid], lm.CommentId)
			}
			lm.Type = content.Loc
			lm.Uid = content.Uid
			lm.Status = notifyvo.MsgStatusNormal
		} else {
			lm.Status = notifyvo.MsgStatusRecalled
		}

		msgs = append(msgs, &lm)
	}

	return msgs, noteLikings, commentLikings
}

func (s *Service) attachUserToMentionMsgs(ctx context.Context, msgs []*notifyentity.MentionedMsg) ([]*dto.MentionMsgWithUser, error) {
	uids := xslice.Uniq(xslice.Extract(msgs, func(m *notifyentity.MentionedMsg) int64 { return m.Uid }))

	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.MentionMsgWithUser, 0, len(msgs))
	for _, msg := range msgs {
		result = append(result, &dto.MentionMsgWithUser{
			MentionedMsg: msg,
			User:         users[msg.Uid],
		})
	}

	return result, nil
}

func (s *Service) attachUserToReplyMsgs(ctx context.Context, msgs []*notifyentity.ReplyMsg) ([]*dto.ReplyMsgWithUser, error) {
	uids := xslice.Uniq(xslice.Extract(msgs, func(m *notifyentity.ReplyMsg) int64 { return m.Uid }))

	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.ReplyMsgWithUser, 0, len(msgs))
	for _, msg := range msgs {
		result = append(result, &dto.ReplyMsgWithUser{
			ReplyMsg: msg,
			User:     users[msg.Uid],
		})
	}

	return result, nil
}

func (s *Service) attachUserToLikesMsgs(ctx context.Context, msgs []*notifyentity.LikesMsg) ([]*dto.LikesMsgWithUser, error) {
	uids := xslice.Uniq(xslice.Extract(msgs, func(m *notifyentity.LikesMsg) int64 { return m.Uid }))

	users, err := s.userAdapter.BatchGetUser(ctx, uids)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.LikesMsgWithUser, 0, len(msgs))
	for _, msg := range msgs {
		result = append(result, &dto.LikesMsgWithUser{
			LikesMsg: msg,
			User:     users[msg.Uid],
		})
	}

	return result, nil
}
