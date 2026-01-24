package comment

import (
	"context"

	v1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
)

type CommentAdapterImpl struct {
	cli v1.CommentServiceClient
}

func NewCommentAdapterImpl(c v1.CommentServiceClient) *CommentAdapterImpl {
	return &CommentAdapterImpl{
		cli: c,
	}
}

var _ repository.CommentAdapter = &CommentAdapterImpl{}

func (c *CommentAdapterImpl) CheckCommented(ctx context.Context,
	p *repository.CheckCommentedParams,
) (*repository.CheckCommentedResult, error) {
	m := make(map[int64]*v1.BatchCheckUserOnObjectRequest_Objects)
	m[p.Uid] = &v1.BatchCheckUserOnObjectRequest_Objects{Oids: p.NoteIds}
	req := &v1.BatchCheckUserOnObjectRequest{
		Mappings: m,
	}
	resp, err := c.cli.BatchCheckUserOnObject(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	pairs := resp.GetResults()
	commented := make(map[int64]bool, len(pairs))
	for _, item := range pairs[p.Uid].GetList() {
		commented[item.GetOid()] = item.GetCommented()
	}

	return &repository.CheckCommentedResult{
		Commented: commented,
	}, nil
}
