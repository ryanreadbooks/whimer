package generics

type DefaultOpter[T any] interface {
	Default() T
}

type Opter[T any] interface {
	~func(*T)
}

func MakeOpt[T DefaultOpter[T], P Opter[T]](opts ...P) *T {
	var opt T
	opt = opt.Default()
	for _, o := range opts {
		o(&opt)
	}

	return &opt
}
