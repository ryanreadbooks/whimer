package svc

import (
	"context"
	"time"

	"github.com/ryanreadbooks/folium/sdk"
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/repo"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	seqerReplyKey = "reply-id-seqer"
)

type CommentSvc struct {
	c     *config.Config
	root  *ServiceContext
	repo  *repo.Repo
	seqer *sdk.Client
}

func NewCommentSvc(ctx *ServiceContext, repo *repo.Repo) *CommentSvc {
	s := &CommentSvc{
		c:    ctx.Config,
		repo: repo,
		root: ctx,
	}

	var err error
	s.seqer, err = sdk.NewClient(sdk.WithGrpc(s.c.ThreeRd.Grpc.Seqer))
	if err != nil {
		panic(err)
	}

	return s
}

func isRootReply(root, parent uint64) bool {
	return root == 0 && parent == 0
}

// 发表评论
func (s *CommentSvc) ReplyAdd(ctx context.Context, req *model.ReplyReq) error {
	var (
		uid      = metadata.GetUid(ctx)
		oid      = req.Oid
		rootId   = req.RootId
		parentId = req.ParentId
		ip       = xnet.IpAsInt(metadata.GetClientIp(ctx))
	)

	replyId, err := s.seqer.GetId(ctx, seqerReplyKey, 10000)
	if err != nil {
		logx.Errorf("reply add get reply id err: %v", err)
		return global.ErrInternal
	}

	now := time.Now().Unix()
	reply := comm.Model{
		Id:       replyId,
		Oid:      oid,
		CType:    int8(req.Type),
		Content:  req.Content,
		Uid:      uid,
		RootId:   rootId,
		ParentId: parentId,
		ReplyUid: req.ReplyUid,
		State:    0, // TODO define state of reply
		Ip:       ip,
		Ctime:    now,
		Mtime:    now,
	}

	err = s.repo.Queue.AddReply(ctx, &reply)
	if err != nil {
		logx.Errorf("push reply to queue err: %v, replyId: %d", err, replyId)
		return global.ErrInternal
	}

	// TODO notify reply_uid

	return nil
}

func (s *CommentSvc) ReplyDel(ctx context.Context) error {

	return nil
}
