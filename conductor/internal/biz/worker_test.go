package biz

import (
	"container/list"
	"context"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"

	. "github.com/smartystreets/goconvey/convey"
)

func newTestWorkerBiz() *WorkerBiz {
	return &WorkerBiz{
		conf: &config.Config{
			WorkerConfig: config.WorkerConfig{
				LongPollTimeout: 5 * time.Second,
			},
		},
		waitingWorkers: make(map[string]*list.List),
	}
}

func TestRemoveWaiting(t *testing.T) {
	Convey("测试 removeWaiting 基本功能", t, func() {
		b := newTestWorkerBiz()
		taskType := "test_task"

		w := &waitingWorker{
			workerId: "worker-1",
			taskType: taskType,
			taskCh:   make(chan *model.Task, 1),
			doneCh:   make(chan struct{}),
		}

		Convey("添加 worker 后队列长度为 1", func() {
			b.addWaiting(w)
			So(b.GetWaitingCount(taskType), ShouldEqual, 1)

			Convey("移除后队列长度为 0", func() {
				b.removeWaiting(w)
				So(b.GetWaitingCount(taskType), ShouldEqual, 0)
				So(w.element, ShouldBeNil)
			})
		})
	})
}

func TestRemoveWaiting_Multiple(t *testing.T) {
	Convey("测试多个 worker 的移除", t, func() {
		b := newTestWorkerBiz()
		taskType := "test_task"

		workers := make([]*waitingWorker, 5)
		for i := 0; i < 5; i++ {
			workers[i] = &waitingWorker{
				workerId: "worker-" + string(rune('0'+i)),
				taskType: taskType,
				taskCh:   make(chan *model.Task, 1),
				doneCh:   make(chan struct{}),
			}
			b.addWaiting(workers[i])
		}

		So(b.GetWaitingCount(taskType), ShouldEqual, 5)

		Convey("移除中间的 worker", func() {
			b.removeWaiting(workers[2])
			So(b.GetWaitingCount(taskType), ShouldEqual, 4)
		})

		Convey("移除第一个 worker", func() {
			b.removeWaiting(workers[0])
			So(b.GetWaitingCount(taskType), ShouldEqual, 4)
		})

		Convey("移除最后一个 worker", func() {
			b.removeWaiting(workers[4])
			So(b.GetWaitingCount(taskType), ShouldEqual, 4)
		})
	})
}

func TestRemoveWaiting_DoubleRemove(t *testing.T) {
	Convey("测试重复移除不 panic", t, func() {
		b := newTestWorkerBiz()
		taskType := "test_task"

		w := &waitingWorker{
			workerId: "worker-1",
			taskType: taskType,
			taskCh:   make(chan *model.Task, 1),
			doneCh:   make(chan struct{}),
		}

		b.addWaiting(w)
		b.removeWaiting(w)

		So(func() { b.removeWaiting(w) }, ShouldNotPanic)
		So(b.GetWaitingCount(taskType), ShouldEqual, 0)
	})
}

func TestRemoveWaiting_NilElement(t *testing.T) {
	Convey("测试 element 为 nil 时移除不 panic", t, func() {
		b := newTestWorkerBiz()

		w := &waitingWorker{
			workerId: "worker-1",
			taskType: "test_task",
			element:  nil,
		}

		So(func() { b.removeWaiting(w) }, ShouldNotPanic)
	})
}

func TestDispatchAfterRemove(t *testing.T) {
	Convey("测试移除后 dispatch 跳过已移除的 worker", t, func() {
		b := newTestWorkerBiz()
		taskType := "test_task"

		workers := make([]*waitingWorker, 3)
		for i := 0; i < 3; i++ {
			workers[i] = &waitingWorker{
				workerId: "worker-" + string(rune('0'+i)),
				taskType: taskType,
				taskCh:   make(chan *model.Task, 1),
				doneCh:   make(chan struct{}),
			}
			b.addWaiting(workers[i])
		}

		Convey("移除第一个 worker 后 dispatch 应发给第二个", func() {
			b.removeWaiting(workers[0])

			task := &model.Task{TaskType: taskType}
			dispatched := b.DispatchTask(context.Background(), task)

			So(dispatched, ShouldBeTrue)

			// worker-1 应该收到任务
			select {
			case <-workers[1].taskCh:
				So(true, ShouldBeTrue)
			case <-time.After(100 * time.Millisecond):
				So(false, ShouldBeTrue) // fail
			}

			// worker-0 不应收到
			select {
			case <-workers[0].taskCh:
				So(false, ShouldBeTrue) // fail
			default:
				So(true, ShouldBeTrue)
			}
		})
	})
}
