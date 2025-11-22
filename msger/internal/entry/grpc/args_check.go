package grpc

import "github.com/ryanreadbooks/whimer/msger/internal/global"

type ChatIdUserIdGetter interface {
	GetChatId() int64
	GetUserId() int64
}

func checkChatIdUserId(g ChatIdUserIdGetter) error {
	if g.GetChatId() <= 0 {
		return global.ErrChatNotExist
	}

	if g.GetUserId() == 0 {
		return global.ErrChatUserEmpty
	}

	return nil
}

type ChatIdMsgIdGetter interface {
	GetChatId() int64
	GetMsgId() int64
}

func checkChatIdMsgId(g ChatIdMsgIdGetter) error {
	if g.GetChatId() <= 0 {
		return global.ErrChatNotExist
	}

	if g.GetMsgId() == 0 {
		return global.ErrMsgNotExist
	}

	return nil
}
