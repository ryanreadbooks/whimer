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
	s.seqer, err = sdk.NewClient(sdk.WithGrpc(s.c.External.Grpc.Seqer))
	if err != nil {
		panic(err)
	}

	return s
}

func isRootReply(root, parent uint64) bool {
	return root == 0 && parent == 0
}

// 发表评论
func (s *CommentSvc) ReplyAdd(ctx context.Context, req *model.ReplyReq) (*model.ReplyRes, error) {
	var (
		uid      = metadata.Uid(ctx)
		oid      = req.Oid
		rootId   = req.RootId
		parentId = req.ParentId
		ip       = xnet.IpAsInt(metadata.ClientIp(ctx))
	)

	replyId, err := s.seqer.GetId(ctx, seqerReplyKey, 10000)
	if err != nil {
		logx.Errorf("reply add get reply id err: %v", err)
		return nil, global.ErrInternal
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
		State:    int8(model.ReplyStateNormal),
		Ip:       ip,
		Ctime:    now,
		Mtime:    now,
	}

	err = s.repo.Bus.AddReply(ctx, &reply)
	if err != nil {
		logx.Errorf("push reply to queue err: %v, replyId: %d", err, replyId)
		return nil, global.ErrInternal
	}

	// TODO notify reply_uid

	return &model.ReplyRes{ReplyId: replyId, Uid: uid}, nil
}

func (s *CommentSvc) ReplyDel(ctx context.Context, rid uint64) error {

	return nil
}
