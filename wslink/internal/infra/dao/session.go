package dao

import (
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type Session struct {
	Id             string              `redis:"id" mapstructure:"id"`
	Uid            int64               `redis:"uid" mapstructure:"uid"` // 连接所属用户
	Device         model.Device        `redis:"device" mapstructure:"device"`
	Status         ws.SessionStatus `redis:"status" mapstructure:"status"`
	Ctime          int64               `redis:"ctime" mapstructure:"ctime"` // 连接建立时间
	LastActiveTime int64               `redis:"last_active_time" mapstructure:"last_active_time"`
	Reside         string              `redis:"reside" mapstructure:"reside"` // 所属的服务ip
	Ip             string              `redis:"ip" mapstructure:"ip"`
}
