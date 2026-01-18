package user

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user/model"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取用户卡片信息
func (h *UserHandler) GetHoverProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req, err := xhttp.ParseValidate[usermodel.HoverReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.userBiz.GetHoverProfile(ctx, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, res)
	}
}
