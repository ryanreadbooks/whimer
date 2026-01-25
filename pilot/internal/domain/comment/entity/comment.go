package entity

// 评论基础信息
type Comment struct {
	Id        int64
	Oid       int64
	Type      int32
	Content   string
	Uid       int64
	RootId    int64
	ParentId  int64
	Ruid      int64
	LikeCount int64
	HateCount int64
	Ctime     int64
	Mtime     int64
	Ip        string
	IsPin     bool
	SubsCount int64
	Images    []*CommentImage
	AtUsers   []*AtUser
}

type AtUser struct {
	Uid      int64
	Nickname string
}

type CommentImage struct {
	Key    string
	Width  uint32
	Height uint32
	Format string
	Type   string
}

// 子评论列表
type SubComments struct {
	Items      []*Comment
	NextCursor int64
	HasNext    bool
}

// 带子评论的评论
type DetailedComment struct {
	Root        *Comment
	SubComments *SubComments
}
