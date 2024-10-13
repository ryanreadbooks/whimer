package slices

// 分段处理 每次返回batchsize数量
func BatchExec[T any](list []T, batchsize int, f func(start, end int) error) error {
	l := len(list)
	if batchsize <= 0 {
		batchsize = l
	}

	total := l / batchsize
	if l%batchsize != 0 {
		total++
	}

	var final error
	for i := 0; i < total; i++ {
		start := i * batchsize
		end := (i + 1) * batchsize
		if end > l {
			end = l
		}
		err := f(start, end)
		if err != nil {
			final = err
			break
		}
	}

	return final
}
