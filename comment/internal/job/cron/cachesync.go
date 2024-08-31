package cron

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	cronv3 "github.com/robfig/cron/v3"
)

type CacheSyncer struct {
	c   *cronv3.Cron
	srv *svc.ServiceContext
}

type logger struct{}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	xlog.Msg(msg).Fields(keysAndValues...).Info()
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	xlog.Msg(msg).Err(err).Fields(keysAndValues...).Error()
}

func NewCacheSyncer(spec string, srv *svc.ServiceContext) (*CacheSyncer, error) {
	c := cronv3.New(cronv3.WithChain(
		cronv3.Recover(&logger{}),
	))

	syncer := &CacheSyncer{
		c:   c,
		srv: srv,
	}
	_, err := c.AddJob(spec, syncer)

	return syncer, err
}

func MustNewCacheSyncer(spec string, srv *svc.ServiceContext) *CacheSyncer {
	s, err := NewCacheSyncer(spec, srv)
	if err != nil {
		panic(err)
	}

	return s
}

func (c *CacheSyncer) Run() {
	xlog.Msg("cache syncer starts running...").Info()
	err := c.srv.CommentSvc.FullSyncReplyCountCache(context.Background())
	if err != nil {
		xlog.Msg("cache syncer full sync failed").Err(err).Error()
	}
	xlog.Msg("cache syncer runs done.").Info()
}

func (c *CacheSyncer) Start() {
	c.c.Start()
}

func (c *CacheSyncer) Stop() {
	c.c.Stop()
}
