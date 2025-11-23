package biz

type Biz struct {
	CommentBiz
	CommentInteractBiz
}

func New() Biz {
	return Biz{
		CommentBiz:         NewCommentBiz(),
		CommentInteractBiz: NewCommentInteractBiz(),
	}
}
