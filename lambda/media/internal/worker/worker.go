package worker

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/config"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/ffmpeg"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/storage"
)

type Worker struct {
	w            *worker.Worker
	storage      *storage.Storage
	videoHandler *VideoHandler
}

func NewWorker(c *config.Config) (*Worker, error) {
	store, err := storage.New(c.Storage)
	if err != nil {
		return nil, fmt.Errorf("init storage failed: %w", err)
	}

	var ffOpts []func(*ffmpeg.FFmpeg)
	if c.FFmpeg.BinPath != "" {
		ffOpts = append(ffOpts, ffmpeg.WithBinPath(c.FFmpeg.BinPath))
	}
	if c.FFmpeg.TempDir != "" {
		ffOpts = append(ffOpts, ffmpeg.WithTempDir(c.FFmpeg.TempDir))
	}
	ff := ffmpeg.New(ffOpts...)

	processor := ffmpeg.NewProcessor(ff, store, c.Video.UseStream)
	videoHandler := NewVideoHandler(processor)

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
	w.RegisterHandler(TaskTypeVideoProcess, videoHandler.Handle)

	return &Worker{
		w:            w,
		storage:      store,
		videoHandler: videoHandler,
	}, nil
}

func (w *Worker) Run(ctx context.Context) error {
	return w.w.Run(ctx)
}

func (w *Worker) Stop() {
	w.w.Stop()
}
