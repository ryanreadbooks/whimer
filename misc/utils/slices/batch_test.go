package slices

import (
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
