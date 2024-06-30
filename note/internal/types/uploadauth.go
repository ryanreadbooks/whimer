package types

import (
	"github.com/ryanreadbooks/whimer/note/internal/global"
)

var (
	uploadResourceAllowed = map[string]struct{}{
		"image": {},
		// "video": 1, //TODO uncomment it when supporting video resource
	}
)

// 请求获取资源上传凭证
type UploadAuthReq struct {
	Resource string `json:"resource" form:"resource"`
	Source   string `json:"source" form:"source,optional"`
	MimeType string `json:"mime" form:"mime"`
}

func (r *UploadAuthReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	_, ok := uploadResourceAllowed[r.Resource]
	if !ok {
		return global.ErrUnsupportedResource
	}

	return nil
}

type UploadAuthResHeaders struct {
	Auth        string `json:"auth"`
	Sha256      string `json:"sha256"`
	Date        string `json:"date"`
	Token       string `json:"token"`
}

// 上传凭证响应
type UploadAuthRes struct {
	FildIds     string               `json:"fild_id"`
	CurrentTime int64                `json:"current_time"`
	ExpireTime  int64                `json:"expire_time"`
	UploadAddr  string               `json:"upload_addr"`
	Headers     UploadAuthResHeaders `json:"headers"`
}
