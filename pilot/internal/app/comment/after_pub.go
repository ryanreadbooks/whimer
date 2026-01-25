package comment

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/dto"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
)

// 评论发布后执行的操作
func (s *Service) AfterCommentPublished(
	ctx context.Context,
	commentId int64,
	cmd *dto.PublishCommentCommand,
) {
	uid := metadata.Uid(ctx)

	// 异步处理
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "commentapp.after_pub",
		Job: func(ctx context.Context) error {
			// 1. 追加最近联系人
			atUsers := cmd.AtUsers.ToMentionVo()
			s.userDomainService.AppendRecentContactsAtUser(ctx, uid, atUsers)

			// 2. TODO 通知被@的用户

			// 3. TODO 通知回复的用户

			// 4. TODO 通知笔记作者
			
			return nil
		},
	})
}
