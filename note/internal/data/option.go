package data

// GetOption 数据获取选项
type GetOption struct {
	UseCache bool // 是否使用缓存获取数据
	SetCache bool // 是否在缓存未命中时回填缓存
}

// GetOptionFunc 选项函数
type GetOptionFunc func(*GetOption)

// DefaultGetOption 默认选项：使用缓存并回填
func DefaultGetOption() *GetOption {
	return &GetOption{
		UseCache: true,
		SetCache: true,
	}
}

// ApplyOptions 应用选项
func ApplyOptions(opts ...GetOptionFunc) *GetOption {
	o := DefaultGetOption()
	for _, fn := range opts {
		fn(o)
	}
	return o
}

// WithUseCache 设置是否使用缓存
func WithUseCache(use bool) GetOptionFunc {
	return func(o *GetOption) {
		o.UseCache = use
	}
}

// WithSetCache 设置是否回填缓存
func WithSetCache(set bool) GetOptionFunc {
	return func(o *GetOption) {
		o.SetCache = set
	}
}

// WithoutCache 不使用缓存
func WithoutCache() GetOptionFunc {
	return func(o *GetOption) {
		o.UseCache = false
		o.SetCache = false
	}
}

// WithCacheOnly 仅使用缓存，不回填
func WithCacheOnly() GetOptionFunc {
	return func(o *GetOption) {
		o.UseCache = true
		o.SetCache = false
	}
}
