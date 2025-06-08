package auth

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	accessv1 "github.com/ryanreadbooks/whimer/passport/api/access/v1"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Auth struct {
	accessv1.AccessServiceClient
}

type Config struct {
	Addr string
}

func New(c *Config) (*Auth, error) {
	cli, err := zrpc.NewClientWithTarget(c.Addr)
	if err != nil {
		return nil, err
	}

	a := &Auth{accessv1.NewAccessServiceClient(cli.Conn())}

	return a, nil
}

func NewFromConn(conn zrpc.Client) *Auth {
	return &Auth{accessv1.NewAccessServiceClient(conn.Conn())}
}

func MustAuther(c xconf.Discovery) *Auth {
	authCli, err := zrpc.NewClient(c.AsZrpcClientConf())
	if err != nil {
		panic(err)
	}
	return NewFromConn(authCli)
}

func rawSignInReq(sessId, platform string) *accessv1.IsCheckedInRequest {
	return &accessv1.IsCheckedInRequest{
		SessId:   sessId,
		Platform: platform,
	}
}

func allReq(sessId string) *accessv1.IsCheckedInRequest {
	return rawSignInReq(sessId, "")
}

func webReq(sessId string) *accessv1.IsCheckedInRequest {
	return rawSignInReq(sessId, "web")
}

func (a *Auth) getCookie(r *http.Request) (sessId string, err error) {
	cookie, err := r.Cookie("WHIMERSESSID")
	if err != nil || len(cookie.Value) == 0 {
		err = xerror.ErrNotLogin
		return
	}

	if a == nil {
		err = xerror.ErrInternal
		logx.Errorf("getCookie auth instance is nil")
		return
	}

	sessId = cookie.Value
	return
}

func (a *Auth) User(ctx context.Context, r *http.Request) (uid int64, sessId string, err error) {
	sessId, err = a.getCookie(r)
	if err != nil {
		return
	}

	resp, err := a.IsCheckedIn(ctx, allReq(sessId))
	if err != nil {
		return
	}

	if !resp.Signed {
		err = xerror.ErrNotLogin
		return
	}

	uid = resp.User.Uid

	return
}

func (a *Auth) UserWeb(ctx context.Context, r *http.Request) (uid int64, sessId string, err error) {
	sessId, err = a.getCookie(r)
	if err != nil {
		return
	}

	resp, err := a.IsCheckedIn(ctx, webReq(sessId))
	if err != nil {
		return
	}

	if !resp.Signed {
		err = xerror.ErrNotLogin
		return
	}

	uid = resp.User.Uid

	return
}
