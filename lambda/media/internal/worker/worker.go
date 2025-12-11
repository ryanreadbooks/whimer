package worker

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/config"
)

type Worker struct {
	w *worker.Worker
}

func NewWorker(c *config.Config) (*Worker, error) {
	opts := worker.Options{
		HostConf:    c.Conductor,
		Concurrency: c.Worker.Concurrency,
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = 2
	}

	w, err := worker.New(opts)
	if err != nil {
		return nil, err
	}

	w.RegisterHandler(TaskTypeImageProcess, HandleImageProcess)
	w.RegisterHandler(TaskTypeVideoProcess, HandleVideoProcess)

	return &Worker{w: w}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	return w.w.Run(ctx)
}

func (w *Worker) Stop() {
	w.w.Stop()
}
