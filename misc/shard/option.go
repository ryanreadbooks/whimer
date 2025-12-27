package shard

type option struct {
	ttlSec                int64 // second
	keepaliveRetryWaitSec int64 // second
}

type Option func(*option)

func WithTTL(ttlSec int64) Option {
	return func(o *option) {
		o.ttlSec = ttlSec
	}
}

func WithKeepaliveRetryWaitSec(sec int64) Option {
	return func(o *option) {
		o.keepaliveRetryWaitSec = sec
	}
}

func defaultOption() *option {
	return &option{
		ttlSec:                10,
		keepaliveRetryWaitSec: 1,
	}
}
