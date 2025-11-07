package whisper

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
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
		return "", xerror.Wrapf(err, "remote create p2p chat failed")
	}

	return resp.ChatId, nil
}

// 创建群聊
func (b *Biz) CreateGroupChat(ctx context.Context) (string, error) {

	return "", fmt.Errorf("not implemented")
}
