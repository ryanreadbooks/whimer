package job

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/srv"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type Syncer struct {
	c   *cron.Cron
	srv *srv.Service
}

func NewSyncer(cfg *config.Config, srv *srv.Service) (*Syncer, error) {
	c := cron.New(cron.WithChain(
		cron.Recover(&xlog.CronLogger{}),
	))

	s := &Syncer{
		c:   c,
		srv: srv,
	}

	_, err := c.AddFunc(cfg.Cron.SyncerSpec, s.SyncSummaryCache)
	if err != nil {
		return nil, err
	}
	_, err = c.AddFunc(cfg.Cron.SummarySpec, s.SyncRecordSummary)
	if err != nil {
		return nil, err
	}

	return s, err
}

func MustNewSyncer(cfg *config.Config, srv *srv.Service) *Syncer {
	if s, err := NewSyncer(cfg, srv); err != nil {
		panic(err)
	} else {
		return s
	}
}

func (s *Syncer) SyncSummaryCache() {
	xlog.Msg("counter syncer starts running...").Info()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	err := s.srv.CounterSrv.CounterBiz.SyncCacheSummary(ctx)
	if err != nil {
		xlog.Msg("syncer sync cache summary failed").Err(err).Error()
	}
	xlog.Msg("syncer sync cache summary done.").Info()
}

func (s *Syncer) SyncRecordSummary() {
	xlog.Msg("counter record summary syncer starts running...").Info()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := s.srv.CounterSrv.CounterBiz.SyncSummaryFromRecords(ctx)
	if err != nil {
		xlog.Msg("counter record summary syncer failed").Err(err).Error()
	}
	xlog.Msg("counter record summary syncer done.").Info()
}

func (c *Syncer) Start() {
	c.c.Start()
}

func (c *Syncer) Stop() {
	c.c.Stop()
	xlog.Msg("counter syncer stopped.").Info()
}
