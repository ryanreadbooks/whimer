package dep

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
)

var (
	// 笔记服务
	noteFeed     notev1.NoteFeedServiceClient
	noteInteract notev1.NoteInteractServiceClient

	// 用户服务
	auther *auth.Auth
	userer userv1.UserServiceClient

	// 评论服务
	commenter commentv1.ReplyServiceClient
	err       error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)
	userer = xgrpc.NewRecoverableClient(c.Backend.Passport, userv1.NewUserServiceClient)

	noteFeed = xgrpc.NewRecoverableClient(c.Backend.Note, notev1.NewNoteFeedServiceClient)
	noteInteract = xgrpc.NewRecoverableClient(c.Backend.Note, notev1.NewNoteInteractServiceClient)
}

func Auther() *auth.Auth {
	return auther
}

func Userer() userv1.UserServiceClient {
	return userer
}

func NoteInteracter() notev1.NoteInteractServiceClient {
	return noteInteract
}

func NoteFeeder() notev1.NoteFeedServiceClient {
	return noteFeed
}

func Commenter() commentv1.ReplyServiceClient {
	return commenter
}