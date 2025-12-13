package shard

import "time"

// 当前实例的分片信息
type Shard struct {
	// 全部实例的key列表
	Keys []string `json:"keys"`

	// 当前实例索引
	Index int `json:"index"`

	// 总实例数量
	Total int `json:"total"`

	// 是否活跃
	Active bool `json:"active"`

	// 当前实例信息的最后更新时间
	LastUpdateAt time.Time `json:"last_update_at"`
}

func (s *Shard) Equal(other *Shard) bool {
	for i := range s.Keys {
		if s.Keys[i] != other.Keys[i] {
			return false
		}
	}
	if s.Index != other.Index {
		return false
	}
	if s.Total != other.Total {
		return false
	}
	if s.Active != other.Active {
		return false
	}
	return s.LastUpdateAt.Equal(other.LastUpdateAt)
}
