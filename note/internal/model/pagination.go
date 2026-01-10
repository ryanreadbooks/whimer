package model

import "math"

const (
	// MaxCursor 游标最大值，用于分页查询时的初始游标
	MaxCursor = math.MaxInt64
)

type PageResult struct {
	NextCursor int64
	HasNext    bool
}

type PageResultV2 struct {
	NextCursor string
	HasNext    bool
}
