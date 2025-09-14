package model

type PageResult struct {
	NextCursor int64
	HasNext    bool
}

type PageResultV2 struct {
	NextCursor string
	HasNext    bool
}
