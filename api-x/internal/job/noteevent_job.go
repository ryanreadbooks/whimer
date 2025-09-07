package job

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"golang.org/x/sync/errgroup"
)

// 定时从redis中捞笔记的互动事件并同步到es中
type NoteEventJobManager struct {
	ctx    context.Context
	cancel context.CancelFunc

	tick      time.Duration
	ticker    *time.Ticker
	searchBiz *bizsearch.SearchBiz
}

type NoteEventJobManagerConfig struct {
	Tick      time.Duration
	NumOfList int
}

// create and start ticker underneath
func NewNoteEventJobManager(opt NoteEventJobManagerConfig, bizz *biz.Biz) *NoteEventJobManager {
	tick := opt.Tick
	ctx, cancel := context.WithCancel(context.Background())
	m := NoteEventJobManager{
		ctx:       ctx,
		cancel:    cancel,
		tick:      tick,
		ticker:    time.NewTicker(tick),
		searchBiz: bizz.SearchBiz,
	}

	return &m
}

func (s *NoteEventJobManager) invoke() error {
	var eg errgroup.Group
	eg.Go(func() error {
		return s.searchBiz.NoteStatSyncer.PollLikeCount(s.ctx)
	})

	eg.Go(func() error {
		return s.searchBiz.NoteStatSyncer.PollCommentCount(s.ctx)
	})

	err := eg.Wait()
	if err != nil {
		xlog.Msgf("nove evnt job mgr invoke err").Err(err).Errorx(s.ctx)
	} else {
		xlog.Msgf("note evnt job mgr invoke done").Infox(s.ctx)
	}

	return err
}

func (s *NoteEventJobManager) Start() {
	for {
		select {
		case <-s.ctx.Done():
			// exit
			xlog.Msg("note event job manager ctx.Done").Info()
			return
		case <-s.ticker.C:
			_ = s.invoke() // ignore error here
		}
	}
}

func (s *NoteEventJobManager) Stop() {
	s.ticker.Stop()
	s.cancel()
}
