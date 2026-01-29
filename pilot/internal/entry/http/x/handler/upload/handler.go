package upload

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter"
	adapterstorage "github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/storage"
)

type Handler struct {
	storageAdapter *adapterstorage.OssRepositoryImpl
}

func NewHandler(c *config.Config) *Handler {
	return &Handler{
		storageAdapter: adapter.StorageAdapter(),
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
		ticket, err := h.storageAdapter.GetUploadTicket(ctx, storagevo.ObjectType(req.Resource), req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetTempCredsResp{
			UploadFile: UploadFile{
				Bucket: ticket.Bucket,
				Ids:    ticket.FileIds,
			},
			UploadCreds: UploadCreds{
				TmpAccessKey: ticket.AccessKey,
				TmpSecretKey: ticket.SecretKey,
				SessionToken: ticket.SessionToken,
				UploadAddr:   ticket.UploadAddr,
				ExpireAt:     ticket.ExpireAt,
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
		ticket, err := h.storageAdapter.GetPostPolicyTicket(ctx, storagevo.ObjectType(req.Resource), req.Sha256, req.MimeType)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &GetPostPolicyCredsResp{
			FileId:     ticket.FileId,
			UploadAddr: ticket.UploadAddr,
			Form:       ticket.Form,
		})
	}
}
