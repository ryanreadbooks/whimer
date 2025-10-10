package websocket

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type NotifyType string

const (
	NotifyUserMsg    NotifyType = "user_msg"
	NotifySysNotice  NotifyType = "sys_notice"
	NotifySysReply   NotifyType = "sys_reply"
	NotifySysMention NotifyType = "sys_mention"
	NotifySysLikes   NotifyType = "sys_likes"
)

// 通知用户被@时下发的数据
type NotifySysContent struct {
	MsgId   string          `json:"msg_id"`
	MsgType model.MsgType   `json:"msg_type"`
	Content json.RawMessage `json:"content"` // 自定义的内容
}

type notifySysContentInner struct {
	Type NotifyType `json:"type"`
	*NotifySysContent
}
