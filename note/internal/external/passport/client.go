package passport

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/passport/sdk/access"
	"github.com/zeromicro/go-zero/zrpc"
)

var (
	Client access.Access
)

func New(c *config.Config) error {
	cli, err := zrpc.NewClientWithTarget(c.ThreeRd.Grpc.Passport)
	if err != nil {
		return err
	}

	Client = access.NewAccess(cli)
	return nil
}

func rawSignInReq(sessId, platform string) *access.CheckSignInReq {
	return &access.CheckSignInReq{
		SessId:   sessId,
		Platform: platform,
	}
}

func allSignInReq(sessId string) *access.CheckSignInReq {
	return rawSignInReq(sessId, "")
}

func webSignInReq(sessId string) *access.CheckSignInReq {
	return rawSignInReq(sessId, "web")
}

func CheckSignIn(ctx context.Context, r *http.Request) (uid uint64, sessId string, err error) {
	cookie, err := r.Cookie("WHIMERSESSID")
	if err != nil || len(cookie.Value) == 0 {
		err = errorx.ErrNotLogin
		return
	}

	if Client == nil {
		err = errorx.ErrInternal
		return
	}

	resp, err := Client.CheckSignIn(ctx, allSignInReq(cookie.Value))
	if err != nil {
		return
	}

	if !resp.Signed {
		err = errorx.ErrNotLogin
		return
	}

	uid = resp.User.Uid
	sessId = cookie.Value // escaped

	return
}
