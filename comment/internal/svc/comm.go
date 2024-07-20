package svc

import "github.com/ryanreadbooks/whimer/comment/internal/repo"

type CommentSvc struct {
	repo *repo.Repo

	Ctx *ServiceContext
}

func NewCommentSvc(ctx *ServiceContext, repo *repo.Repo) *CommentSvc {
	return &CommentSvc{
		repo: repo,
		Ctx:  ctx,
	}
}
