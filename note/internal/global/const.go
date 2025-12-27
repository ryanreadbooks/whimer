package global

// 计数服务的业务码
const (
	NoteLikeBizcode int32 = 20001 + iota // 点赞
)

// conductor调度
const (
	NoteProcessNamespace = "note-process"
)

const (
	NoteImageProcessTaskType = "image_process"
	NoteVideoProcessTaskType = "video_process"
)
