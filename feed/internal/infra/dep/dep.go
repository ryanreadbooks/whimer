package dep

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
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

	// 用户关系服务
	relationer relationv1.RelationServiceClient
	err        error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)

	userer = xgrpc.NewRecoverableClient(c.Backend.Passport,
		userv1.NewUserServiceClient, func(nc userv1.UserServiceClient) { userer = nc })

	noteFeed = xgrpc.NewRecoverableClient(c.Backend.Note,
		notev1.NewNoteFeedServiceClient, func(nc notev1.NoteFeedServiceClient) { noteFeed = nc })

	noteInteract = xgrpc.NewRecoverableClient(c.Backend.Note,
		notev1.NewNoteInteractServiceClient, func(nc notev1.NoteInteractServiceClient) { noteInteract = nc })

	commenter = xgrpc.NewRecoverableClient(c.Backend.Comment,
		commentv1.NewReplyServiceClient, func(nc commentv1.ReplyServiceClient) { commenter = nc })

	relationer = xgrpc.NewRecoverableClient(c.Backend.Relation,
		relationv1.NewRelationServiceClient, func(rsc relationv1.RelationServiceClient) { relationer = rsc })
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

func Relationer() relationv1.RelationServiceClient {
	return relationer
}
