package comment

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/dto"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	notifyvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
)

// 评论发布后执行的操作
func (s *Service) AfterCommentPublished(
	ctx context.Context,
	commentId int64,
	cmd *dto.PublishCommentCommand,
) {
	operator := metadata.Uid(ctx)

	// 异步处理
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "pilot.commentapp.after_pub.handle_at_users",
		Job: func(ctx context.Context) error {
			// 1. 追加最近联系人
			atUsers := cmd.AtUsers.ToMentionVo()
			// 去重+排除自己
			validAtUsers := make(mentionvo.AtUserList, 0, len(atUsers))
			for _, atUser := range atUsers {
				if atUser.Uid == operator {
					continue
				}
				validAtUsers = append(validAtUsers, atUser)
			}

			validAtUsers = xslice.UniqF(validAtUsers, func(v *mentionvo.AtUser) int64 { return v.Uid })

			if len(validAtUsers) > 0 {
				err := s.userDomainService.AppendRecentContactsAtUser(ctx, operator, validAtUsers)
				// 2. TODO 通知被@的用户
				err2 := s.afterCommentPublishedHandleAtUsers(ctx, operator, validAtUsers, commentId, cmd)
				err3 := errors.Join(err, err2)
				return err3
			}

			return nil
		},
	})

	// 3. 通知被回复的用户
	s.afterCommentPublishedHandleReplyUser(ctx, operator, commentId, cmd)

	// 4. TODO 通知笔记作者
}

func (s *Service) afterCommentPublishedHandleAtUsers(
	ctx context.Context,
	operator int64,
	targetUsers []*mentionvo.AtUser,
	commentId int64,
	cmd *dto.PublishCommentCommand,
) error {
	err := s.systemNotifyDomainService.NotifyAtUsersOnComment(ctx, &notifyvo.NotifyAtUsersOnCommentParam{
		Uid:         operator,
		TargetUsers: targetUsers,
		Content: &notifyvo.NotifyAtUsersOnCommentParamContent{
			SourceUid: operator,
			Comment:   cmd.Content,
			NoteId:    cmd.Oid,
			CommentId: commentId,
			RootId:    cmd.RootId,
			ParentId:  cmd.ParentId,
		},
	})
	if err != nil {
		return xerror.Wrapf(err, "system notify service notify at users on comment failed").WithCtx(ctx)
	}

	return nil
}

func (s *Service) afterCommentPublishedHandleReplyUser(
	ctx context.Context,
	operator int64,
	commentId int64,
	cmd *dto.PublishCommentCommand,
) {
	var (
		loc             notifyvo.NotifyMsgLocation
		targetCommentId int64
	)

	if cmd.PubOnOidDirectly() {
		loc = notifyvo.NotifyMsgOnNote
	} else {
		loc = notifyvo.NotifyMsgOnComment
		targetCommentId = cmd.ParentId
	}

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "pilot.commentapp.after_pub.handle_reply_user",
		Job: func(ctx context.Context) error {
			comment, err := s.commentAdapter.GetComment(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "get comment failed").WithCtx(ctx)
			}

			if comment.Uid == operator { // 自己回复自己不需要处理
				return nil
			}

			atUsers := make([]*mentionvo.AtUser, 0, len(comment.AtUsers))
			for _, atUser := range comment.AtUsers {
				atUsers = append(atUsers, &mentionvo.AtUser{
					Uid:      atUser.Uid,
					Nickname: atUser.Nickname,
				})
			}

			commentContent := &notifyvo.CommentContent{
				Text:    comment.Content,
				AtUsers: atUsers,
			}

			content, err := json.Marshal(commentContent)
			if err != nil {
				return xerror.Wrapf(err, "json marshal comment content failed").WithCtx(ctx)
			}

			err = s.systemNotifyDomainService.NotifyUserReply(ctx, &notifyvo.NotifyUserReplyParam{
				Loc:            loc,
				TriggerComment: commentId,
				TargetComment:  targetCommentId,
				SrcUid:         operator,
				RecvUid:        cmd.ReplyUid,
				NoteId:         cmd.Oid,
				Content:        content,
			})
			if err != nil {
				return xerror.Wrapf(err, "system notify service notify user reply failed").WithCtx(ctx)
			}

			return nil
		},
	})
}
