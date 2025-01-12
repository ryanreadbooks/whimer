package model

import (
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/global"
)

var (
	uploadResourceAllowed = map[string]struct{}{
		"image": {},
		// "video": 1, //TODO uncomment it when supporting video resource
	}
)

const (
	maxRequiredUpload = 8
)

// 请求获取资源上传凭证
type UploadAuthRequest struct {
	Resource string `json:"resource" form:"resource"`
	Source   string `json:"source" form:"source,optional"`
	Count    int32  `json:"count" form:"count"`
}

func (r *UploadAuthRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	_, ok := uploadResourceAllowed[r.Resource]
	if !ok {
		return global.ErrUnsupportedResource
	}

	if r.Count > maxRequiredUpload {
		return global.ErrArgs.Msg("最多上传8张图片")
	}

	return nil
}

type UploadAuthResponseHeaders struct {
	Auth   string `json:"auth"`
	Sha256 string `json:"sha256"`
	Date   string `json:"date"`
	Token  string `json:"token"`
}

func (h *UploadAuthResponseHeaders) AsPb() *notev1.UploadAuthResponseHeaders {
	return &notev1.UploadAuthResponseHeaders{
		Auth:   h.Auth,
		Sha256: h.Sha256,
		Date:   h.Date,
		Token:  h.Token,
	}
}

// 上传凭证响应
type UploadAuthResponse struct {
	FildId      string                    `json:"fild_id"`
	CurrentTime int64                     `json:"current_time"`
	ExpireTime  int64                     `json:"expire_time"`
	UploadAddr  string                    `json:"upload_addr"`
	Headers     UploadAuthResponseHeaders `json:"headers"`
}

func (r *UploadAuthResponse) AsPb() *notev1.GetUploadAuthResponse {
	return &notev1.GetUploadAuthResponse{
		FileId:      r.FildId,
		CurrentTime: r.CurrentTime,
		ExpireTime:  r.ExpireTime,
		UploadAddr:  r.UploadAddr,
		Headers:     r.Headers.AsPb(),
	}
}
