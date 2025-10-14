package pushcmd

// websocket推送action定义
type Action string

const (
	ActionPullP2P     = "pull_p2p"
	ActionPullUnreads = "pull_unreads"
)
