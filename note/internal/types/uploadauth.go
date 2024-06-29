package types

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/global"
)

var (
	uploadResourceAllowed = map[string]int{
		"image": 9,
		// "video": 1, //TODO uncomment it when supporting video resource
	}
)

// 请求获取资源上传凭证
type UploadAuthReq struct {
	Resource string `json:"resource" form:"resource"`
	Count    int    `json:"count" form:"count"`
	Source   string `json:"source" form:"source,optional"`
}

func (r *UploadAuthReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	maxCount, ok := uploadResourceAllowed[r.Resource]
	if !ok {
		return global.ErrUnsupportedResource
	}

	if r.Count <= 0 {
		return global.ErrArgs.Msg("资源数量不对")
	}

	if r.Count > maxCount {
		return global.ErrArgs.Msg(fmt.Sprintf("最多上传%d个资源", maxCount))
	}

	return nil
}

// 上传凭证响应
type UploadAuthRes struct {
	FildIds     []string `json:"fild_ids"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}
