package upload

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
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

// 获取临时凭证
func (h *Handler) GetTemporaryCreds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateForm[GetTempCredsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		creds, err := h.storageBiz.GetUploadTemporaryTicket(
			ctx,
			&bizstorage.GetUploadTemporaryTicketRequest{
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

func (h *Handler) GetPostPolicyCreds() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidateForm[GetPostPolicyCredsReq](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.storageBiz.GetPostPolicyUploadTicket(
			ctx,
			&bizstorage.GetPostPolicyUploadTicketRequest{
				Resource: uploadresource.Type(req.Resource),
				Sha256:   req.Sha256,
				Size:     req.Size,
				MimeType: req.MimeType,
			})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
