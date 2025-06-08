package dao

import (
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type Session struct {
	Id             string           `redis:"id" mapstructure:"id"`
	Uid            int64            `redis:"uid" mapstructure:"uid"` // 连接所属用户
	Device         model.Device     `redis:"device" mapstructure:"device"`
	Status         ws.SessionStatus `redis:"status" mapstructure:"status"`
	Ctime          int64            `redis:"ctime" mapstructure:"ctime"` // 连接建立时间
	LastActiveTime int64            `redis:"last_active_time" mapstructure:"last_active_time"`
	LocalIp        string           `redis:"local_ip" mapstructure:"local_ip"` // 所属的服务ip
	Ip             string           `redis:"ip" mapstructure:"ip"`
}

func (s *Session) GetId() string {
	return s.Id
}

func (s *Session) SetId(c string) {
	s.Id = c
}

func (s *Session) GetRemote() string {
	return s.Ip
}

func (s *Session) GetLocalIp() string {
	return s.LocalIp
}

func (s *Session) GetDevice() model.Device {
	return s.Device
}

func (s *Session) IsActive() bool {
	return s.Status == ws.StatusActive
}
