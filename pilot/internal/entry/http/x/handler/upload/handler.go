package upload

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/storage"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"
)

type Handler struct {
	storageBiz *bizstorage.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		storageBiz: bizz.UploadBiz,
	}
}

func (h *Handler) GetTempCreds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateForm[GetTempCredsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		creds, err := h.storageBiz.RequestUploadTemporaryTicket(ctx,
			bizstorage.RequestUploadTemporaryTicket{
				Resource: uploadresource.Type(req.Resource),
				Source:   req.Source,
				Count:    req.Count,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetTempCredsResp{
			UploadFile: UploadFile{
				Bucket: creds.Bucket,
				Ids:    creds.FileIds,
			},
			UploadCreds: UploadCreds{
				TmpAccessKey: creds.AccessKey,
				TmpSecretKey: creds.SecretKey,
				SessionToken: creds.SessionToken,
				UploadAddr:   creds.UploadAddr,
				ExpireAt:     creds.ExpireAt,
			},
		})
	}
}

type Req struct {
	Bucket   string `form:"bucket"`
	Key      string `form:"key"`
	NumBytes int32  `form:"num_bytes,optional"`
}

