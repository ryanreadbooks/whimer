package utils

func Must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func Must1[T any](v T, err error) T {
	Must(err)
	return v
}

func Must2[T any, P any](v1 T, v2 P, err error) (T, P) {
	Must(err)
	return v1, v2
}
