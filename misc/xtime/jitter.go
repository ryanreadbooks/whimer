package xtime

import (
	"math/rand"
	"time"
)

// 返回随机过期时间
func Jitter(d time.Duration) int64 {
	sec := time.Duration(rand.Int63n(int64(d.Seconds())))
	if rand.Intn(2) == 0 {
		return -int64(sec)
	}
	return int64(sec)
}

func JitterDuration(d time.Duration) time.Duration {
	sec := Jitter(d)
	return time.Second * time.Duration(sec)
}
