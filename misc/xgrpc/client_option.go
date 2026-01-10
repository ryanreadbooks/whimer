package xgrpc

type clientOption struct {
	withoutDefaultInterceptor bool
}

type ClientOption func(*clientOption)

func WithoutDefaultInterceptor() ClientOption {
	return func(o *clientOption) {
		o.withoutDefaultInterceptor = true
	}
}

func defaultClientOption() *clientOption {
	return &clientOption{
		withoutDefaultInterceptor: false,
	}
}
