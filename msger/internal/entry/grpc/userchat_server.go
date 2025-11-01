package grpc

import (
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
)

type UserChatServiceServer struct {
	pbuserchat.UnimplementedUserChatServiceServer
	
	Srv *srv.Service
}


func NewUserChatServiceServer(svc *srv.Service) *UserChatServiceServer {
	return &UserChatServiceServer{
		Srv: svc,
	}
}