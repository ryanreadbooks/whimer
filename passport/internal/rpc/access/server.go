package access

import (
	"context"
	"net/url"

	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/passport/internal/model/platform"
	"github.com/ryanreadbooks/whimer/passport/internal/model/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	"github.com/ryanreadbooks/whimer/passport/sdk/access"
	"github.com/ryanreadbooks/whimer/passport/sdk/user"
	
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccessServer struct {
	access.UnimplementedAccessServer

	Svc *svc.ServiceContext
}

func NewAccessServer(s *svc.ServiceContext) *AccessServer {
	return &AccessServer{
		Svc: s,
	}
}

// 检查是否有登录
func (s *AccessServer) CheckSignIn(ctx context.Context, in *access.CheckSignInReq) (*access.CheckSignInRes, error) {
	if len(in.SessId) == 0 {
		return nil, status.Error(codes.PermissionDenied, "empty sessid")
	}

	var (
		res    access.CheckSignInRes
		meInfo *profile.MeInfo
		err    error
	)

	val, err := url.PathUnescape(in.SessId)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "invalid sessid")
	}
	in.SessId = val

	if len(in.Platform) == 0 {
		meInfo, err = s.Svc.AccessSvc.CheckSignedIn(ctx, in.SessId)
		if err != nil {
			if errorx.IsInternal(err) {
				return nil, status.Error(codes.Internal, err.Error())
			}
		}
	} else {
		if !platform.Supported(in.Platform) {
			return nil, status.Error(codes.PermissionDenied, "unsupported platform")
		}
		meInfo, err = s.Svc.AccessSvc.CheckPlatformSignedIn(ctx, in.SessId, in.Platform)
		if err != nil {
			if errorx.IsInternal(err) {
				return nil, status.Error(codes.Internal, err.Error())
			}
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
