package kafka

import (
	"context"
	"errors"
	"io"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type BatchReaderConfig struct {
	BatchSize    int           // 每批次获取的消息数量
	BatchTimeout time.Duration // 获取每批次消息的最长等待时间
}

// 批量读取kafka消息
type BatchReader struct {
	r       *kafka.Reader
	c       BatchReaderConfig
	mu      sync.Mutex
	msgs    chan kafka.Message
	errChan chan error
	ctx     context.Context
	cancel  context.CancelFunc
	closed  bool
	errOnce sync.Once
	err     error
}

func NewBatchReader(r *kafka.Reader, c BatchReaderConfig) *BatchReader {
	ctx, cancel := context.WithCancel(context.Background())
	if c.BatchSize == 0 {
		c.BatchSize = 100
	}
	
	br := &BatchReader{
		r:       r,
		c:       c,
		msgs:    make(chan kafka.Message, r.Config().QueueCapacity),
		errChan: make(chan error, 1),
		ctx:     ctx,
		cancel:  cancel,
	}

	// 启动后台goroutine持续获取消息
	go br.prefetchMessages()

	return br
}

// 后台预取消息
func (r *BatchReader) prefetchMessages() {
	defer close(r.msgs)
	defer close(r.errChan)

	for {
		select {
		case <-r.ctx.Done():
			return
		default:
			msg, err := r.r.FetchMessage(r.ctx)
			if err != nil {
				// 如果是上下文取消，不视为错误
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}

				select {
				case r.errChan <- err:
				default:
				}
				return
			}

			// 发送消息到通道
			select {
			case r.msgs <- msg:
			case <-r.ctx.Done():
				return
			}
		}
	}
}

// 批量获取消息
func (r *BatchReader) BatchFetchMessages(ctx context.Context) ([]kafka.Message, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil, io.EOF
	}

	if r.err != nil {
		return nil, r.err
	}

	var batch []kafka.Message
	timer := time.NewTimer(r.c.BatchTimeout)
	defer timer.Stop()

	for {
		select {
		case msg, ok := <-r.msgs:
			if !ok {
				select {
				case err := <-r.errChan:
					r.errOnce.Do(func() {
						r.err = err
					})
					return batch, err
				default:
					r.err = io.EOF
					return batch, io.EOF
				}
			}

			batch = append(batch, msg)
			if len(batch) >= r.c.BatchSize {
				return batch, nil
			}

		case err := <-r.errChan:
			r.errOnce.Do(func() {
				r.err = err
			})
			return batch, err

		case <-timer.C:
			if len(batch) > 0 {
				return batch, nil
			}
			// 如果没有消息，继续等待
			timer.Reset(r.c.BatchTimeout)

		case <-ctx.Done():
			return batch, ctx.Err()
		}
	}
}

// 提交消息
func (r *BatchReader) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return io.EOF
	}

	return r.r.CommitMessages(ctx, msgs...)
}

func (r *BatchReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true
	r.cancel()
	return r.r.Close()
}
