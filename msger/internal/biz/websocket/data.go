package websocket

import ()

// 通知的命令类型
type NotifyCmd string

const (
	NotifyUserMsgCmd    NotifyCmd = "peer_msg_incoming"
	NotifySysNoticeCmd  NotifyCmd = "sys_notice"
	NotifySysReplyCmd   NotifyCmd = "sys_reply"
	NotifySysMentionCmd NotifyCmd = "sys_mention"
	NotifySysLikesCmd   NotifyCmd = "sys_likes"
)

// 通知客户端需要执行的动作
type NotifyAction string

const (
	NotifyActionPullUnread NotifyAction = "pull_unread" // 拉取最新的未读数
)

// 下发的动作
type NotifySysCmdData struct {
	Cmd    NotifyCmd      `json:"cmd"`
	Action []NotifyAction `json:"actions"`
}
