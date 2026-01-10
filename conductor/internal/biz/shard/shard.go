package shard

import (
	"encoding/json"
	"fmt"
)

const (
	// 分片总大小
	sizeOfShard = 1024
)

// Strategy 分片策略
type Strategy interface {
	// Calculate 根据实例数量计算分片划分
	Calculate(totalSize int, instanceCount int) []Range
}

// EvenStrategy 均匀分片策略
type EvenStrategy struct{}

func (s *EvenStrategy) Calculate(totalSize int, instanceCount int) []Range {
	if instanceCount <= 0 {
		return nil
	}

	shards := make([]Range, instanceCount)
	baseSize := totalSize / instanceCount
	remainder := totalSize % instanceCount

	start := 0
	for i := range instanceCount {
		size := baseSize
		// 前 remainder 个分片多分一个
		if i < remainder {
			size++
		}
		shards[i] = Range{
			ShardId: i,
			Start:   start,
			End:     start + size,
		}
		start += size
	}

	return shards
}

// Range 分片范围
type Range struct {
	// 分片 ID
	ShardId int `json:"shard_id"`

	// 起始值（包含）
	Start int `json:"start"`

	// 结束值（不包含）
	End int `json:"end"`
}

func (s Range) String() string {
	return fmt.Sprintf("shard-%d:[%d,%d)", s.ShardId, s.Start, s.End)
}

// Value etcd 中存储的分片值
type Value struct {
	// 持有者实例 ID
	Holder string `json:"holder"`

	Range
}

func (v *Value) Marshal() string {
	data, _ := json.Marshal(v)
	return string(data)
}

func unmarshalValue(data []byte) (*Value, error) {
	var v Value
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return &v, nil
}
