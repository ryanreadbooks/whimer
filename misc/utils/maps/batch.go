package maps

import (
	stderr "errors"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
)

func decideMapBatches[M ~map[K]V, K comparable, V any](m M, batchsize int) []M {
	var batches []M
	lm := len(m)
	if batchsize <= 0 {
		batchsize = lm
	}

	keys := Keys(m)
	for i := 0; i < lm; i += batchsize {
		end := i + batchsize
		if end > lm {
			end = lm
		}
		target := keys[i:end]
		batchMap := make(M, len(target))
		for _, k := range target {
			batchMap[k] = m[k]
		}
		batches = append(batches, batchMap)
	}

	return batches
}

// 遍历map并且分批执行指定函数 每次返回batchsize个元素
func BatchExec[M ~map[K]V, K comparable, V any](m M, batchsize int, fn func(target M) error) error {
	for _, batch := range decideMapBatches(m, batchsize) {
		err := fn(batch)
		if err != nil {
			return err
		}
	}

	return nil
}

// wg will wait inside the function
func BatchAsyncExec[M ~map[K]V, K comparable, V any](wg *sync.WaitGroup, m M, batchsize int, fn func(target M) error) error {
	var (
		batches = decideMapBatches(m, batchsize)
		errors  = make(chan error, len(batches))
	)

	for _, batch := range batches {
		wg.Add(1)
		concurrent.SafeGo(func() {
			defer wg.Done()
			errors <- fn(batch)
		})
	}

	wg.Wait() // wait inside the batch exec function
	var finals []error
	for err := range errors {
		if err != nil {
			finals = append(finals, err)
		}
	}

	return stderr.Join(finals...)
}
