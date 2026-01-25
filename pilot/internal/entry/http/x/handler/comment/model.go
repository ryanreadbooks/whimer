package comment

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
)

type UploadTicket struct {
	StoreKeys   []string `json:"store_keys"`
	CurrentTime int64    `json:"current_time"`
	ExpireTime  int64    `json:"expire_time"`
	UploadAddr  string   `json:"upload_addr"`
	Token       string   `json:"token"`
}

func newUploadTicket(t *dep.UploadTicket) *UploadTicket {
	return &UploadTicket{
		StoreKeys:   t.FileIds,
		CurrentTime: t.CurrentTime,
		ExpireTime:  t.ExpireTime,
		UploadAddr:  t.UploadAddr,
		Token:       t.Token,
	}
}
