package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	bizuser "github.com/ryanreadbooks/whimer/api-x/internal/biz/user"
	usermodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xslice"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserHandler struct {
	userBiz *bizuser.UserBiz
}

func NewUserHandler(c *config.Config, bizz *biz.Biz) *UserHandler {
	return &UserHandler{
		userBiz: bizz.UserBiz,
	}
}

func (h *UserHandler) ListInfos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[usermodel.ListInfosReq](httpx.ParseForm, r)
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
		resp, err := h.userBiz.ListUsers(ctx, uids)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *UserHandler) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[usermodel.GetUserReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.userBiz.GetUser(ctx, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 获取用户粉丝列表
func (h *UserHandler) ListUserFans() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[usermodel.GetFanOrFollowingListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
		)

		resp, total, err := h.userBiz.ListUserFans(ctx, req.Uid, req.Page, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &usermodel.GetFanOrFollowingListResp{
			Items: resp,
			Total: total,
		})
	}
}

// 获取用户关注列表
func (h *UserHandler) ListUserFollowings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[usermodel.GetFanOrFollowingListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
		)

		resp, total, err := h.userBiz.ListUserFollowings(ctx, req.Uid, req.Page, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &usermodel.GetFanOrFollowingListResp{
			Items: resp,
			Total: total,
		})
	}
}

// 获取用户的投稿数量、点赞数量等信息
func (h *UserHandler) GetUserStat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[usermodel.HoverReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		var (
			ctx = r.Context()
		)

		stat, err := h.userBiz.GetUserStat(ctx, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &stat)
	}
}

func (h *UserHandler) GetAllSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := h.userBiz.GetSettings(r.Context())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
