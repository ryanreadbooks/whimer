package model

type ListResult[T comparable] struct {
	NextCursor T
	HasNext    bool
}
