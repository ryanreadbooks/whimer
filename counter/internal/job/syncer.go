package job

import (
	"context"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/ryanreadbooks/whimer/counter/internal/svc"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type Syncer struct {
	c   *cron.Cron
	srv *svc.ServiceContext
}

func NewSyncer(spec string, srv *svc.ServiceContext) (*Syncer, error) {
	c := cron.New(cron.WithChain(
		cron.Recover(&xlog.CronLogger{}),
	))

	s := &Syncer{
		c:   c,
		srv: srv,
	}

	_, err := c.AddJob(spec, s)

	return s, err
}

func MustNewSyncer(spec string, srv *svc.ServiceContext) *Syncer {
	if s, err := NewSyncer(spec, srv); err != nil {
		panic(err)
	} else {
		return s
	}
}

func (s *Syncer) Run() {
	xlog.Msg("counter syncer starts running...").Debug()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()
	err := s.srv.RecordSvc.SyncCacheSummary(ctx)
	if err != nil {
		xlog.Msg("syncer sync cache summary failed").Err(err).Error()
	}
	xlog.Msg("syncer sync cache summary done.").Info()
}

func (c *Syncer) Start() {
	c.c.Start()
}

func (c *Syncer) Stop() {
	c.c.Stop()
}
