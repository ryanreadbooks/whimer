package generics

type Integer interface {
	UnSignedInteger | SignedInteger
}

type UnSignedInteger interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type SignedInteger interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Float interface {
	~float32 | ~float64
}

type StringOrNumber interface {
	~string | Float | Integer
}
