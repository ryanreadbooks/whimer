package xretry

import (
	"math"
	"sync"
	"time"
)

type ExponentialBackoff struct {
	mu              sync.Mutex
	initialInterval time.Duration
	maxInterval     time.Duration
	multiplier      float64
	currentRetry    int
	maxRetries      int
}

// NewBackoff创建指数退避实例
// 参数说明：
// - initial: 初始等待时间（必须>0）
// - maxInterval: 时间间隔上限（0表示无上限）
// - multiplier: 指数增长倍数（通常2.0）
// - maxRetries: 最大重试次数（-1表示无限重试）
func NewBackoff(initial, maxInterval time.Duration, multiplier float64, maxRetries int) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialInterval: initial,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
		maxRetries:      maxRetries,
	}
}

func NewPermenantBackoff(initial, maxInterval time.Duration, multiplier float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialInterval: initial,
		maxInterval:     maxInterval,
		multiplier:      multiplier,
		maxRetries:      -1,
	}
}

// 重试3次
func NewDefaultBackoff(initial time.Duration) *ExponentialBackoff {
	return &ExponentialBackoff{
		initialInterval: initial,
		multiplier:      2.0,
		maxRetries:      3,
	}
}

// NextBackOff返回下一次等待时间
//
// 第二个参数返回false不应该继续等待
func (b *ExponentialBackoff) NextBackOff() (time.Duration, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 检查是否达到重试次数上限
	if b.maxRetries >= 0 && b.currentRetry >= b.maxRetries {
		return 0, false
	}

	interval := time.Duration(float64(b.initialInterval) * math.Pow(b.multiplier, float64(b.currentRetry)))

	// 应用时间间隔上限（当maxInterval>0时生效）
	if b.maxInterval > 0 && interval > b.maxInterval {
		interval = b.maxInterval
	}

	b.currentRetry++
	return interval, true
}

func (b *ExponentialBackoff) Success() {
	b.Reset()
}

func (b *ExponentialBackoff) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.currentRetry = 0
}

func (b *ExponentialBackoff) GetCurrentRetry() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.currentRetry
}
