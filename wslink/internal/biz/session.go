package biz

import (
	"container/heap"
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	cxmap "github.com/ryanreadbooks/whimer/misc/concurrent/xmap"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xrand"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"

	"github.com/google/uuid"
)

const (
	sessionTTL = 30 // 30 seconds
)

type UnSendableSession interface {
	GetId() string
	SetId(string)
	GetDevice() model.Device
	GetRemote() string
	GetLocalIp() string
}

type sessionNoSend struct {
	*dao.Session
}

var _ = (UnSendableSession)(*&sessionNoSend{})

type Session interface {
	UnSendableSession
	Close(context.Context)
	Send(ctx context.Context, data []byte) error
}

type SessionBiz interface {
	// Connect will create a session
	Connect(ctx context.Context, conn Session) error
	// Disconnect will invalidate a session
	Disconnect(ctx context.Context, cid string) error
	// Heartbeat will renew a session
	Heartbeat(ctx context.Context, conn Session) error
	// close will make all session go offline
	Close(ctx context.Context)

	// local sessions
	GetSessionByUid(ctx context.Context, uid int64) ([]Session, error)
	GetSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]Session, error)
	// local sessions and forwarded sessions
	GetUnSendSessionByUid(ctx context.Context, uid int64) ([]UnSendableSession, error)
	GetUnSendSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]UnSendableSession, error)

	RespectivelyGetSessionByUid(ctx context.Context, uid int64) ([]Session, []UnSendableSession, error)
	// 按照sessIds批量获取，分开本机和非本机
	RespectivelyGetSessionById(ctx context.Context, sessIds []string) ([]Session, []UnSendableSession, error)
}

type sessionBiz struct {
	closing atomic.Bool
	closed  chan struct{}

	// 和本机建立的连接
	sessions     *cxmap.ShardedMap[string, Session]
	sessionsHeap *SessionCountdownQueue
}

func NewSessionBiz() SessionBiz {
	b := &sessionBiz{
		sessions:     cxmap.NewShardedMap[string, Session](config.Conf.System.ConnShard),
		sessionsHeap: NewSessionCountdownQueue(),
		closed:       make(chan struct{}, 1),
	}

	b.closing.Store(false)
	b.keepAliveSession()

	return b
}

func (b *sessionBiz) keepAliveSession() {
	concurrent.SafeGo(func() {
		ticker := time.NewTicker(time.Second * 1) // 一秒pop一次
		defer ticker.Stop()

		for {
			select {
			case <-b.closed:
				return
			case now := <-ticker.C:
				b.doKeepAlive(now)
			}
		}
	})
}

func (b *sessionBiz) doKeepAlive(now time.Time) {
	// pop from max heap
	if b.sessionsHeap.Len() > 0 {
		cdt := heap.Pop(b.sessionsHeap).(*SessionCountdown)
		if cdt == nil {
			xlog.Msg("session countdown is nil").Error()
		} else {
			if cdt.NextTime > now.Unix()+5 && b.sessions.Has(cdt.Id) {
				heap.Push(b.sessionsHeap, cdt) // 还没到时间，重新放回
				return
			}

			err := infra.Dao().SessionDao.SetTTL(context.Background(), cdt.Id, sessionTTL)
			if err != nil {
				xlog.Msgf("session %s set ttl failed", cdt.Id).Err(err).Error()
			}
			// put back to heap if id is still valid
			if b.sessions.Has(cdt.Id) {
				sec := xrand.Range(10, 20) // 10-20秒后继续续期
				cdt.NextTime = now.Add(time.Second * time.Duration(sec)).Unix()
				heap.Push(b.sessionsHeap, cdt)
			}
		}
	}
}

func (b *sessionBiz) beginKeepAlive(ctx context.Context, sess Session) error {
	err := infra.Dao().SessionDao.SetTTL(ctx, sess.GetId(), sessionTTL)
	if err != nil {
		return xerror.Wrapf(err, "session dao failed to set ttl").WithExtras("id", sess.GetId())
	}

	sec := xrand.Range(10, 20)
	heap.Push(b.sessionsHeap, &SessionCountdown{
		Id:       sess.GetId(),
		NextTime: time.Now().Add(time.Second * time.Duration(sec)).Unix(),
	})

	return nil
}

func pickSessionByStatus(ss []*dao.Session, st ws.SessionStatus, dev model.Device, uid int64) *dao.Session {
	for _, s := range ss {
		if s.Status == st && s.Device == dev && s.Uid == uid {
			// we take this one
			return s
		}
	}

	return nil
}

// 建立连接
func (b *sessionBiz) Connect(ctx context.Context, conn Session) error {
	var (
		uid    = metadata.Uid(ctx)
		device = conn.GetDevice()
	)

	// we try to find some noactive sessions to re-use it
	curSessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "dao failed to get session by uid").WithExtras("uid", uid).WithCtx(ctx)
	}

	var target *dao.Session = pickSessionByStatus(curSessions, ws.StatusPending, device, uid)
	if target == nil {
		target = pickSessionByStatus(curSessions, ws.StatusNoActive, device, uid)
	}

	now := time.Now().Unix()
	var ds = dao.Session{
		Uid:            uid,
		Device:         device,
		Status:         ws.StatusActive,
		Ctime:          now,
		LastActiveTime: now,
		LocalIp:        conn.GetLocalIp(),
		Ip:             conn.GetRemote(),
	}
	if target == nil {
		// create a new one
		if conn.GetId() == "" {
			newConnId := uuid.NewString()
			conn.SetId(newConnId)
		}
		ds.Id = conn.GetId()
	} else {
		// re-use target id
		ds.Id = target.Id
		conn.SetId(target.Id)
	}

	err = infra.Dao().SessionDao.Create(ctx, &ds)
	if err != nil {
		return xerror.Wrapf(err, "dao failed to create session").
			WithExtras("uid", uid).WithCtx(ctx)
	}

	// connected
	b.sessions.Put(conn.GetId(), conn)
	// 开始续期
	if err := b.beginKeepAlive(ctx, conn); err != nil {
		xlog.Msgf("begin keepalive failed for conn %s", conn.GetId()).Err(err).Errorx(ctx)
	}

	return nil
}

// disconnect a session
func (b *sessionBiz) Disconnect(ctx context.Context, cid string) error {
	if b.closing.Load() {
		return nil
	}

	err := infra.Dao().SessionDao.UpdateStatus(ctx, cid, ws.StatusNoActive)
	if err != nil {
		return xerror.Wrapf(err, "dao failed to update status").
			WithExtras("cid", cid).
			WithCtx(ctx)
	}

	b.sessions.Remove(cid)

	return nil
}

// make cids sessions go offline
func (b *sessionBiz) Close(ctx context.Context) {
	cids := b.sessions.Keys()

	b.closing.Store(true)
	b.closed <- struct{}{}

	var wg sync.WaitGroup
	err := xslice.BatchAsyncExec(&wg, cids, 50, func(start, end int) error {
		for _, cid := range cids[start:end] {
			err := infra.Dao().SessionDao.UpdateStatus(ctx, cid, ws.StatusPending)
			if err != nil {
				xlog.Msgf("session offline dao update status failed").Extras("cid", cid).Err(err).Error()
			}
			b.sessions.Get(cid).Close(ctx)
		}

		return nil
	})

	if err != nil {
		xlog.Msgf("session offline action err").Err(err).Error()
	}

	b.sessions.Clear()
}

func (b *sessionBiz) Heartbeat(ctx context.Context, conn Session) error {
	// heartbeat
	return b.updateSessionLastActiveTime(ctx, conn)
}

func (b *sessionBiz) updateSessionLastActiveTime(ctx context.Context, conn Session) error {
	err := infra.Dao().SessionDao.UpdateLastActiveTime(ctx, conn.GetId(), time.Now().Unix())
	if err != nil {
		return xerror.Wrapf(err, "session dao failed to update heartbeat time").
			WithExtras("cid", conn.GetId()).
			WithCtx(ctx)
	}

	return nil
}

func (b *sessionBiz) GetSessionByUid(ctx context.Context, uid int64) ([]Session, error) {
	logExts := []any{"uid", uid}

	sessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "session dao get by uid failed").WithExtras(logExts...).WithCtx(ctx)
	}

	var result []Session = make([]Session, 0, len(sessions))
	for _, sess := range sessions {
		if sess.Status == ws.StatusActive {
			result = append(result, b.sessions.Get(sess.Id))
		}
	}

	return result, nil
}

func (b *sessionBiz) GetSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]Session, error) {
	candidates, err := b.GetSessionByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get session by uid failed").
			WithExtras("uid", uid, "device", device).
			WithCtx(ctx)
	}

	// filter device
	results := make([]Session, 0, len(candidates))
	for _, c := range candidates {
		if c.GetDevice() == device {
			results = append(results, c)
		}
	}

	return results, nil
}

func (b *sessionBiz) GetUnSendSessionByUid(ctx context.Context, uid int64) ([]UnSendableSession, error) {
	logExts := []any{"uid", uid}

	sessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "session dao get by uid failed").WithExtras(logExts...).WithCtx(ctx)
	}

	var result = make([]UnSendableSession, 0, len(sessions))
	for _, s := range sessions {
		if s.Status == ws.StatusActive {
			result = append(result, s)
		}
	}

	return result, nil
}

func (b *sessionBiz) GetUnSendSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]UnSendableSession, error) {
	candidates, err := b.GetUnSendSessionByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get session by uid failed").
			WithExtras("uid", uid, "device", device).
			WithCtx(ctx)
	}

	// filter device
	results := make([]UnSendableSession, 0, len(candidates))
	for _, c := range candidates {
		if c.GetDevice() == device {
			results = append(results, c)
		}
	}

	return results, nil
}

func (s *sessionBiz) RespectivelyGetSessionByUid(ctx context.Context, uid int64) ([]Session, []UnSendableSession, error) {
	logExts := []any{"uid", uid}

	sessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "session dao get by uid failed").WithExtras(logExts...).WithCtx(ctx)
	}

	local, unsend := s.seperateLocalAndNonLocal(sessions)

	return local, unsend, nil
}

func (s *sessionBiz) seperateLocalAndNonLocal(sessions []*dao.Session) ([]Session, []UnSendableSession) {
	local := make([]Session, 0, len(sessions))
	unlocal := make([]UnSendableSession, 0, len(sessions))
	for _, sess := range sessions {
		if sess != nil && sess.Status == ws.StatusActive {
			if sess.LocalIp == config.GetIpAndPort() && s.sessions.Has(sess.Id) {
				local = append(local, s.sessions.Get(sess.Id))
			} else {
				unlocal = append(unlocal, sess)
			}
		}
	}

	return local, unlocal
}

func (b *sessionBiz) RespectivelyGetSessionById(ctx context.Context, sessIds []string) ([]Session, []UnSendableSession, error) {
	sessions, err := infra.Dao().SessionDao.BatchGetById(ctx, sessIds)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "session failed to batch get by id").WithExtras("ids", sessIds).WithCtx(ctx)
	}

	local, nonlocal := b.seperateLocalAndNonLocal(xmap.Values(sessions))
	return local, nonlocal, nil
}
