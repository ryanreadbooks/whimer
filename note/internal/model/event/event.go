package event

type EventType string

// note事件定义
const (
	// 用户成功发布一篇新笔记 创建或者更新
	NotePublished EventType = "note.published"

	// 用户删除一篇笔记 包含主动删除和笔记可见范围转为私有
	NoteDeleted EventType = "note.deleted"

	// 笔记发布前处理失败（审核不通过/资源处理失败等）
	NoteRejected EventType = "note.rejected"

	// 笔记违规封禁
	NoteBanned EventType = "note.banned"

	// 笔记被点赞
	NoteLiked EventType = "note.liked"

	//  笔记被评论
	NoteCommented EventType = "note.commented"
)

// 笔记成功发布事件
type NotePublishedEventData struct {
	Note *Note `json:"note"`
}

type NoteDeleteReason string

const (
	NoteDeleteReasonPureDelete    NoteDeleteReason = "pure_delete"
	NoteDeleteReasonPrivacyChange NoteDeleteReason = "privacy_change"
)

// 笔记删除事件
type NoteDeletedEventData struct {
	Note   *Note            `json:"note"`
	Reason NoteDeleteReason `json:"reason"`
}

// TODO 笔记审核拒绝事件
type NoteRejectedEventData struct{}

// TODO 笔记被封禁事件
type NoteBannedEventData struct{}

// 笔记点赞/取消点赞事件
type NoteLikedEventData struct {
	NoteId  int64 `json:"note_id"`
	UserId  int64 `json:"user_id"`
	OwnerId int64 `json:"owner_id"`
	IsLiked bool  `json:"is_liked"`
}

type NoteEvent struct {
	Type      EventType `json:"type"`      // 事件类型
	NoteId    string    `json:"note_id"`   // 笔记id 可用于对外暴露为混淆字符串类型
	Timestamp int64     `json:"timestamp"` // 事件时间戳 unix milisecond
	Payload   any       `json:"payload"`   // 事件payload
}
