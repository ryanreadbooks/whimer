package model

import "hash/fnv"

const (
	SizeOfShard = 1024
)

func CalculateShardHash(taskType string) int {
	h := fnv.New64a()
	h.Write([]byte(taskType))
	return int(h.Sum64() % SizeOfShard)
}