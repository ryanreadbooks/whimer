package xrand

import (
	"testing"
)

func TestRange(t *testing.T) {
	t.Log(Range(0, 1))
	t.Log(Range(0, 2))
	t.Log(Range(0, 4))
	t.Log(Range(0, 5))
	t.Log(Range(0, 6))
}
