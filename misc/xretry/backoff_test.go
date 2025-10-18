package xretry

import (
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	b := NewBackoff(time.Second, time.Second*10, 1.3, 8)

	for range 10 {
		wait, ok := b.NextBackOff()
		t.Log(wait.Seconds(), ok)
	}

	t.Log("----------")

	b = NewBackoff(time.Second, time.Second*10, 2, -1)
	for i := range 21 {
		wait, ok := b.NextBackOff()
		t.Log(wait.Seconds(), ok)

		if i == 15 {
			b.Reset()
		}

		if i == 20 {
			break
		}
	}
	t.Log("----------")

	b = NewBackoff(time.Second, 0, 2, -1)
	for range 20 {
		t.Log(b.NextBackOff())
	}

	b = NewDefaultBackoff(time.Millisecond * 300)
	for range 5 {
		t.Log(b.NextBackOff())
	}
}
