package userchat

import "github.com/ryanreadbooks/whimer/msger/internal/model"

type MsgContent interface {
	Content() ([]byte, error)
	MsgType() model.MsgType
	Preview() string
}

// 消息内容详细定义

type MsgContentText struct{}

