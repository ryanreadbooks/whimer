package xrand

import "math/rand"

func Range(start, end int) int {
	if start > end {
		start, end = end, start
	}
	return rand.Intn(end-start+1) + start
}
