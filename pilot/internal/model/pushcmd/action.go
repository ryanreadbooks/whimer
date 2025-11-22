package pushcmd

// websocket推送action定义
type Action string

const (
	ActionPullWhisper = "pull_whisper"
	ActionPullUnreads = "pull_unreads"
)
