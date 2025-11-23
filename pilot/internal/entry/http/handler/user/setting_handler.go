package user

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	usermodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
)

func (h *UserHandler) SetNoteShowSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateJsonBody[SetNoteShowSettingsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		err = h.userBiz.SetNoteShowSettings(ctx, uid, &usermodel.IntegralNoteShowSetting{
			ShowNoteLikes: req.ShowNoteLikes,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
