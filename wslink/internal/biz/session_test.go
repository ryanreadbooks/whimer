package biz

import (
	"container/heap"
	"testing"
)

func TestHeap(t *testing.T) {
	q := NewSessionCountdownQueue()

	heap.Init(q)
	heap.Push(q, &SessionCountdown{
		NextTime: 100,
	})
	heap.Push(q, &SessionCountdown{
		NextTime: 788,
	})
	heap.Push(q, &SessionCountdown{
		NextTime: 23,
	})
	heap.Push(q, &SessionCountdown{
		NextTime: 1,
	})
	heap.Push(q, &SessionCountdown{
		NextTime: 98,
	})
	heap.Push(q, &SessionCountdown{
		NextTime: 23,
	})

	item := heap.Pop(q).(*SessionCountdown)
	heap.Push(q, item)
	t.Log(item)
	t.Log("---------------")

	for q.Len() > 0 {
		item := heap.Pop(q).(*SessionCountdown)
		t.Log(item)
	}

	q2 := NewSessionCountdownQueue()
	i := heap.Pop(q2)
	t.Log(i)

}
