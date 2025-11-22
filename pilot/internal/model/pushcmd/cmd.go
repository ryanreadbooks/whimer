package pushcmd

import "encoding/json"

// websocket推送cmd定义
type Cmd string

const (
	CmdWhisperMsgNotify = "whisper_notify"
	CmdSysMsgNotify     = "sys_notify"
)

type CmdAction struct {
	Cmd     Cmd      `json:"cmd"`
	Actions []Action `json:"actions"`
}

func NewCmdAction(cmd Cmd, action Action, actions ...Action) CmdAction {
	return CmdAction{
		Cmd:     cmd,
		Actions: append([]Action{action}, actions...),
	}
}

func (c CmdAction) Bytes() []byte {
	b, _ := json.Marshal(c)
	return b
}
