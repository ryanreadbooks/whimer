package http

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func GetMyProfileHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signedUser := model.CtxGetUserInfo(r.Context())
		if signedUser == nil {
			xhttp.Error(r, w, global.ErrNotCheckedIn)
			return
		}

		info, err := serv.UserSrv.GetUser(r.Context(), signedUser.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, info)
	}
}

func UpdateMyProfileHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signedUser := model.CtxGetUserInfo(r.Context())
		if signedUser == nil {
			xhttp.Error(r, w, global.ErrNotCheckedIn)
			return
		}

		req, err := xhttp.ParseValidate[model.UpdateUserRequest](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if signedUser.Uid != req.Uid {
			httpx.Error(w, global.ErrPermDenied)
			return
		}

		me, err := serv.UserSrv.UpdateUser(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, me)
	}
}

// 通过服务上传头像
// 头像大小较小可以通过服务器中转
func UpdateMyAvatarHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("avatar")
		if err != nil {
			httpx.Error(w, global.ErrAvatarNotFound)
			return
		}

		req, err := model.ParseAvatarFile(file, header)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		avatarUrl, err := serv.UserSrv.UpdateUserAvatar(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &model.UploadUserAvatarResponse{Url: avatarUrl})
	}
}
