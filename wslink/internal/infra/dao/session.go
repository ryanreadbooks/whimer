package dao

import "github.com/ryanreadbooks/whimer/wslink/internal/model"

type Session struct {
	Id             string              `json:"id"`
	Uid            int64               `json:"uid"` // 连接所属用户
	Device         model.Device        `json:"device,omitempty"`
	Status         model.SessionStatus `json:"status"`
	Ctime          int64               `json:"ctime"`            // 连接建立时间
	LastActiveTime int64               `json:"last_active_time"` // 上次心跳时间
	Reside         string              `json:"reside"`           // 所属的服务ip
	Ip             string              `json:"ip,omitempty"`
}
