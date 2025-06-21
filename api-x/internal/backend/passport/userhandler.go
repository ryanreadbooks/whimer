package passport

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/infra"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserHandler struct{}

func NewUserHandler(c *config.Config) *UserHandler {
	return &UserHandler{}
}

type ListInfosReq struct {
	Uids string `form:"uids"` // 多个用,分隔
}

func (h *UserHandler) ListInfos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ListInfosReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		uidStrs := strings.Split(req.Uids, ",")
		if len(uidStrs) == 0 {
			xhttp.Error(r, w, xerror.ErrArgs)
			return
		}

		uids := make([]int64, 0, len(uidStrs))
		for _, us := range uidStrs {
			uid, err := strconv.ParseInt(us, 10, 64)
			if err == nil {
				uids = append(uids, uid)
			}
		}

		ctx := r.Context()
		uids = xslice.Uniq(uids)
		resp, err := infra.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
			Uids: uids,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp.GetUsers())
	}
}
