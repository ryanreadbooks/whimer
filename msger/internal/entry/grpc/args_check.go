package grpc

import (
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

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

func OrderFromListChatMsgsRequest(in *pbuserchat.ListChatMsgsRequest) (model.Order, error) {
	switch in.GetOrder() {
	case pbuserchat.ListChatMsgsRequest_ORDER_UNSPECIFIED:
		return model.OrderUnspecified, nil
	case pbuserchat.ListChatMsgsRequest_ORDER_ASC:
		return model.OrderAsc, nil
	case pbuserchat.ListChatMsgsRequest_ORDER_DESC:
		return model.OrderDesc, nil
	}
	return model.OrderUnspecified, global.ErrArgs.Msg("invalid order")
}
