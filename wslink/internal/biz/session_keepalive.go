package biz

import "sync"

type SessionCountdown struct {
	Id       string
	NextTime int64
}

type SessionCountdownQueue struct {
	sync.RWMutex
	queue []*SessionCountdown
}

func NewSessionCountdownQueue() *SessionCountdownQueue {
	return &SessionCountdownQueue{
		queue: make([]*SessionCountdown, 0, 16),
	}
}

// implements container/heap.Interface
func (q *SessionCountdownQueue) Len() int {
	q.RLock()
	defer q.RUnlock()

	return len(q.queue)
}

func (q *SessionCountdownQueue) Less(i, j int) bool {
	q.RLock()
	defer q.RUnlock()

	return q.queue[i].NextTime < q.queue[j].NextTime
}

func (q *SessionCountdownQueue) Swap(i, j int) {
	if i < 0 || j < 0 {
		return
	}

	q.Lock()
	defer q.Unlock()
	q.queue[i], q.queue[j] = q.queue[j], q.queue[i]
}

func (q *SessionCountdownQueue) Push(x any) {
	item := x.(*SessionCountdown)
	q.Lock()
	defer q.Unlock()
	q.queue = append(q.queue, item)
}

func (q *SessionCountdownQueue) Pop() any {
	q.Lock()
	defer q.Unlock()

	old := q.queue
	n := len(old)
	if n == 0 {
		return nil
	}

	item := old[n-1]
	old[n-1] = nil
	q.queue = old[0 : n-1]

	return item
}
