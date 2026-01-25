package user

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	appuser "github.com/ryanreadbooks/whimer/pilot/internal/app/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/user/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type UserHandler struct {
	userApp *appuser.Service
}

func NewUserHandler(c *config.Config, manager *app.Manager) *UserHandler {
	return &UserHandler{
		userApp: manager.UserApp,
	}
}

func (h *UserHandler) ListInfos() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.ListUsersReq](httpx.ParseForm, r)
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
		resp, err := h.userApp.ListUsers(ctx, uids)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *UserHandler) GetUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.GetUserReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.userApp.GetUser(ctx, req.Uid)
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
		req, err := xhttp.ParseValidate[dto.GetFanOrFollowingListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.userApp.ListUserFans(ctx, req.Uid, req.Page, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 获取用户关注列表
func (h *UserHandler) ListUserFollowings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.GetFanOrFollowingListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.userApp.ListUserFollowings(ctx, req.Uid, req.Page, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 获取用户的投稿数量、点赞数量等信息
func (h *UserHandler) GetUserStat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.HoverReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		ctx := r.Context()

		stat, err := h.userApp.GetUserStat(ctx, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, stat)
	}
}

func (h *UserHandler) GetAllSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := h.userApp.GetSettings(r.Context())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *UserHandler) AtUserCandidates() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.MentionUserReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		result, err := h.userApp.GetMentionUserCandidates(ctx, uid, req.Search)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, result)
	}
}

// 获取用户卡片信息
func (h *UserHandler) GetHoverProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		req, err := xhttp.ParseValidate[dto.HoverReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.userApp.GetHoverProfile(ctx, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, res)
	}
}

func (h *UserHandler) SetNoteShowSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[dto.SetNoteShowSettingReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		err = h.userApp.SetNoteShowSettings(ctx, uid, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
