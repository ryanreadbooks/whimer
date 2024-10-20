package svc

import (
	"context"
	"testing"

	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/repo"
)

var (
	testRepo = repo.New(config.ConfigForTest())
	ctx      = context.TODO()
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestCounterSvc_SyncSummaryFromRecords(t *testing.T) {
	svc := &CounterSvc{repo: testRepo}
	svc.SyncSummaryFromRecords(ctx)
}
