package vo

// 评论类型
type CommentType int32

const (
	CommentTypeText      CommentType = 0 // 纯文本
	CommentTypeImageText CommentType = 1 // 图文
)

// 点赞/点踩操作
type ThumbAction uint8

const (
	ThumbActionUndo ThumbAction = 0 // 取消
	ThumbActionDo   ThumbAction = 1 // 执行
)

// 置顶操作
type PinAction int8

const (
	PinActionUnpin PinAction = 0 // 取消置顶
	PinActionPin   PinAction = 1 // 置顶
)

// 排序类型
type SortType int32

const (
	SortByTime SortType = 0 // 按时间
	SortByHot  SortType = 1 // 按热度
)
