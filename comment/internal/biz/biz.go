package biz

type Biz struct {
	CommentBiz
	CommentInteractBiz

	AssetManagerBiz *AssetManagerBiz
}

func New() Biz {
	return Biz{
		CommentBiz:         NewCommentBiz(),
		CommentInteractBiz: NewCommentInteractBiz(),
		AssetManagerBiz:    NewAssetManagerBiz(),
	}
}
