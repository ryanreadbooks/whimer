package grpc

import (
	"context"
	"net/url"

	"github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"
	access "github.com/ryanreadbooks/whimer/passport/sdk/access/v1"
	user "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
)

type AccessServiceServer struct {
	access.UnimplementedAccessServiceServer

	Svc *srv.Service
}

func NewAccessServiceServer(s *srv.Service) *AccessServiceServer {
	return &AccessServiceServer{
		Svc: s,
	}
}

// 检查是否有登录
func (s *AccessServiceServer) IsCheckedIn(ctx context.Context, in *access.IsCheckedInRequest) (*access.IsCheckedInResponse, error) {
	if len(in.SessId) == 0 {
		return nil, global.ErrPermDenied.Msg("empty sessId")
	}

	var (
		res    access.IsCheckedInResponse
		meInfo *model.UserInfo
		err    error
	)

	val, err := url.PathUnescape(in.SessId)
	if err != nil {
		return nil, global.ErrPermDenied.Msg("invalid sessId")
	}
	in.SessId = val

	if len(in.Platform) == 0 {
		meInfo, err = s.Svc.AccessSrv.IsCheckedIn(ctx, in.SessId)
		if err != nil {
			return nil, err
		}
	} else {
		if !model.SupportedPlatform(in.Platform) {
			return nil, global.ErrPermDenied.Msg("unsupported platform")
		}
		meInfo, err = s.Svc.AccessSrv.IsCheckedInOnPlatform(ctx, in.SessId, in.Platform)
		if err != nil {
			return nil, err
		}
	}

	if err != nil {
		res.Msg = err.Error()
	}

	if meInfo != nil {
		res.User = &user.UserInfo{
			Uid:       meInfo.Uid,
			Nickname:  meInfo.Nickname,
			Avatar:    meInfo.Avatar,
			StyleSign: meInfo.StyleSign,
			Gender:    meInfo.Gender,
		}
		res.Signed = true
	}

	return &res, nil
}
