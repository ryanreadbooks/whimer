package auth

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	ppac "github.com/ryanreadbooks/whimer/passport/sdk/access/v1"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Auth struct {
	ppac.AccessClient
}

type Config struct {
	Addr string
}

func New(c *Config) (*Auth, error) {
	cli, err := zrpc.NewClientWithTarget(c.Addr)
	if err != nil {
		return nil, err
	}

	a := &Auth{ppac.NewAccessClient(cli.Conn())}

	return a, nil
}

func NewFromConn(conn zrpc.Client) *Auth {
	return &Auth{ppac.NewAccessClient(conn.Conn())}
}

func MustAuther(c xconf.Discovery) *Auth {
	authCli, err := zrpc.NewClient(c.AsZrpcClientConf())
	if err != nil {
		panic(err)
	}
	return NewFromConn(authCli)
}

func rawSignInReq(sessId, platform string) *ppac.CheckSignInReq {
	return &ppac.CheckSignInReq{
		SessId:   sessId,
		Platform: platform,
	}
}

func allReq(sessId string) *ppac.CheckSignInReq {
	return rawSignInReq(sessId, "")
}

func webReq(sessId string) *ppac.CheckSignInReq {
	return rawSignInReq(sessId, "web")
}

func (a *Auth) getCookie(r *http.Request) (sessId string, err error) {
	cookie, err := r.Cookie("WHIMERSESSID")
	if err != nil || len(cookie.Value) == 0 {
		err = errorx.ErrNotLogin
		return
	}

	if a == nil {
		err = errorx.ErrInternal
		logx.Errorf("getCookie auth instance is nil")
		return
	}

	sessId = cookie.Value
	return
}

func (a *Auth) User(ctx context.Context, r *http.Request) (uid uint64, sessId string, err error) {
	sessId, err = a.getCookie(r)
	if err != nil {
		return
	}

	resp, err := a.CheckSignIn(ctx, allReq(sessId))
	if err != nil {
		return
	}

	if !resp.Signed {
		err = errorx.ErrNotLogin
		return
	}

	uid = resp.User.Uid

	// TODO csrf check

	return
}

func (a *Auth) UserWeb(ctx context.Context, r *http.Request) (uid uint64, sessId string, err error) {
	sessId, err = a.getCookie(r)
	if err != nil {
		return
	}

	resp, err := a.CheckSignIn(ctx, webReq(sessId))
	if err != nil {
		return
	}

	if !resp.Signed {
		err = errorx.ErrNotLogin
		return
	}

	uid = resp.User.Uid

	// TODO csrf check

	return
}
