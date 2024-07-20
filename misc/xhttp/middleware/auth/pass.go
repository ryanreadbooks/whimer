package auth

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/errorx"
	ppac "github.com/ryanreadbooks/whimer/passport/sdk/access"
	"github.com/zeromicro/go-zero/zrpc"
)

type Auth struct {
	ppac.Access
}

type Config struct {
	Addr string
}

func New(c *Config) (*Auth, error) {
	cli, err := zrpc.NewClientWithTarget(c.Addr)
	if err != nil {
		return nil, err
	}

	a := &Auth{ppac.NewAccess(cli)}

	return a, nil
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
