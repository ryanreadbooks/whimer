package xslice

import (
	"sync"
	"testing"
)

func TestBatchExec(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	BatchExec(a, 10, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 1, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")

	BatchExec(a, 2, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")

	BatchExec(a, 3, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 4, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 5, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 6, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 7, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")

	BatchExec(a, -1, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
	BatchExec(a, 0, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
	t.Log("============")
}

func TestBatchAsyncExec(t *testing.T) {
	a := []int{1, 2, 3, 4, 5, 6, 7}
	var wg sync.WaitGroup
	BatchAsyncExec(&wg, a, 2, func(start, end int) error {
		t.Log(a[start:end])
		return nil
	})
}

func TestBatchAsyncExec2(t *testing.T) {
	var (
		wg   sync.WaitGroup
		ints []int
		mu   sync.Mutex
	)
	a := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 0}
	BatchAsyncExec(&wg, a, 1, func(start, end int) error {
		mu.Lock()
		defer mu.Unlock()
		ints = append(ints, a[start:end]...)

		return nil
	})

	t.Log(ints)
}
