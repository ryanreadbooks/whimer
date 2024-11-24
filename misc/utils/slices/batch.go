package slices

import (
	stderr "errors"
	"sync"
)

func cal(l int, batchsize int) (batch, total int) {
	batch = batchsize
	if batch <= 0 {
		batch = l
	}

	total = l / batch
	if l%batch != 0 {
		total++
	}

	return
}

// 分段处理 每次返回batchsize数量
// 同步处理
func BatchExec[T any](list []T, batchsize int, f func(start, end int) error) error {
	l := len(list)
	if l == 0 {
		return nil
	}

	batchsize, total := cal(l, batchsize)

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

// 分段异步处理
// 每次处理batchsize的数量
func BatchAsyncExec[T any](wg *sync.WaitGroup, list []T, batchsize int, f func(start, end int) error) error {
	l := len(list)
	if l == 0 {
		return nil
	}
	batchsize, total := cal(l, batchsize)

	errors := make(chan error, total)
	for i := 0; i < total; i++ {
		start := i * batchsize
		end := (i + 1) * batchsize
		if end > l {
			end = l
		}

		wg.Add(1)
		go func(start, end int) {
			defer func() {
				wg.Done()
				if e := recover(); e != nil {
					return
				}
			}()

			err := f(start, end)
			if err != nil {
				errors <- err
			}
		}(start, end)
	}

	wg.Wait()
	close(errors)

	var finals []error
	for err := range errors {
		if err != nil {
			finals = append(finals, err)
		}
	}

	return stderr.Join(finals...)
}
