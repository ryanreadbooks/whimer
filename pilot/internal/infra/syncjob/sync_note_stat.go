package syncjob

import (
	"context"
	"time"

	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/note"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"golang.org/x/sync/errgroup"
)

var instance *NoteStatSyncJob

func GetNoteStatSyncJob() *NoteStatSyncJob {
	return instance
}

func InitNoteStatSyncJob(opt NoteStatSyncJobConfig,
	searchDocCli searchv1.DocumentServiceClient,
	noteStatStore *notecache.StatStore,
) *NoteStatSyncJob {
	instance = NewNoteStatSyncJob(opt, searchDocCli, noteStatStore)
	return instance
}

type NoteStatSyncJob struct {
	ctx    context.Context
	cancel context.CancelFunc

	tick   time.Duration
	ticker *time.Ticker

	searchDocCli  searchv1.DocumentServiceClient
	noteStatStore *notecache.StatStore
}
type NoteStatSyncJobConfig struct {
	Tick      time.Duration
	NumOfList int
}

func NewNoteStatSyncJob(opt NoteStatSyncJobConfig, searchDocCli searchv1.DocumentServiceClient,
	noteStatStore *notecache.StatStore,
) *NoteStatSyncJob {
	tick := opt.Tick
	ctx, cancel := context.WithCancel(context.Background())
	j := &NoteStatSyncJob{
		ctx:           ctx,
		cancel:        cancel,
		tick:          tick,
		ticker:        time.NewTicker(tick),
		searchDocCli:  searchDocCli,
		noteStatStore: noteStatStore,
	}

	return j
}

func (s *NoteStatSyncJob) invoke() error {
	var eg errgroup.Group
	eg.Go(func() error {
		return s.PollLikeCount(s.ctx)
	})

	eg.Go(func() error {
		return s.PollCommentCount(s.ctx)
	})

	err := eg.Wait()
	if err != nil {
		xlog.Msgf("note stat sync job invoke err").Err(err).Errorx(s.ctx)
	} else {
		// xlog.Msgf("note evnt job mgr invoke done").Infox(s.ctx)
	}

	return err
}

func (s *NoteStatSyncJob) Start() {
	concurrent.SafeGo(func() {
		for {
			select {
			case <-s.ctx.Done():
				// exit
				xlog.Msg("note stat sync job ctx.Done").Info()
				return
			case <-s.ticker.C:
				_ = s.invoke() // ignore error here
			}
		}
	})
}

func (s *NoteStatSyncJob) Stop() {
	s.ticker.Stop()
	s.cancel()
}

// consume note like count event
func (s *NoteStatSyncJob) PollLikeCount(ctx context.Context) error {
	stats, err := s.noteStatStore.ConsumeLikeCount(ctx, 1)
	if err != nil {
		return xerror.Wrapf(err, "note stat syncer failed to poll like count").WithCtx(ctx)
	}

	// xlog.Msgf("note stat poll like count len = %d", len(stats)).Debugx(ctx)

	// remove duplicates
	reqs := s.removeDupAndDoMap(stats)
	if len(reqs) != 0 {
		_, err := s.searchDocCli.BatchUpdateNoteLikeCount(ctx,
			&searchv1.BatchUpdateNoteLikeCountRequest{Counts: reqs})
		if err != nil {
			return xerror.Wrapf(err, "note stat syncer update note like count failed").WithCtx(ctx)
		}
	} else {
		// xlog.Msg("note stat poll like count result empty").Debugx(ctx)
	}

	return nil
}

// consume note comment count event
func (s *NoteStatSyncJob) PollCommentCount(ctx context.Context) error {
	stats, err := s.noteStatStore.ConsumeCommentCount(ctx, 1)
	if err != nil {
		return xerror.Wrapf(err, "note stat syncer failed to poll comment count").WithCtx(ctx)
	}

	// xlog.Msgf("note stat poll comment count len = %d", len(stats)).Debugx(ctx)

	reqs := s.removeDupAndDoMap(stats)
	if len(reqs) != 0 {
		_, err := s.searchDocCli.BatchUpdateNoteCommentCount(ctx,
			&searchv1.BatchUpdateNoteCommentCountRequest{Counts: reqs})
		if err != nil {
			return xerror.Wrapf(err, "note stat syncer update note comment count failed").WithCtx(ctx)
		}
	} else {
		// xlog.Msg("note stat poll comment count result empty").Debugx(ctx)
	}

	return nil
}

func (s *NoteStatSyncJob) removeDupAndDoMap(stats []notecache.NoteStatRepr) map[string]int64 {
	tmp := make(map[string]int64, len(stats))
	for _, stat := range stats {
		tmp[stat.NoteId] += stat.Inc
	}

	res := make([]notecache.NoteStatRepr, 0, len(stats))
	for noteId, incr := range tmp {
		if incr != 0 { // 0 means updatign to es is unnecessary
			res = append(res, notecache.NoteStatRepr{
				NoteId: noteId,
				Inc:    incr,
			})
		}
	}

	reqs := make(map[string]int64, len(res))
	for _, s := range res {
		reqs[s.NoteId] = s.Inc
	}

	return reqs
}
