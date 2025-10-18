package biz

// 获取评论信息一些可选项
type getCommentOpt struct {
	populateImages bool
	populateExt    bool
}

type GetCommentOption func(*getCommentOpt)

func defaultGetCommentOption() *getCommentOpt {
	return &getCommentOpt{
		populateImages: true,
		populateExt:    true,
	}
}

func DoNotPopulateImages() GetCommentOption {
	return func(gco *getCommentOpt) {
		gco.populateImages = false
	}
}

func DoNotPopulateExt() GetCommentOption {
	return func(gco *getCommentOpt) {
		gco.populateExt = false
	}
}

func makeGetCommentOption(opts ...GetCommentOption) *getCommentOpt {
	opt := defaultGetCommentOption()
	for _, o := range opts {
		o(opt)
	}

	return opt
}
