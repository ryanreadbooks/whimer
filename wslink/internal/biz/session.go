package biz

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/concurrent/xmap"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type ISession interface {
	GetId() string
	SetId(string)
	GetDevice() model.Device
	GetRemote() string
	Close(context.Context)
	Send(ctx context.Context, data []byte) error
	GetLocalIp() string
}

type SessionBiz interface {
	// Connect will create a session
	Connect(ctx context.Context, conn ISession) error
	// Disconnect will invalidate a session
	Disconnect(ctx context.Context, cid string) error
	// Heartbeat will renew a session
	Heartbeat(ctx context.Context, conn ISession) error
	// OfflineSessions will put all sessions with cids into Offline status for future recovery
	OfflineAllSessions(ctx context.Context)

	GetSessionByUid(ctx context.Context, uid int64) ([]ISession, error)
	GetSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]ISession, error)
}

type sessionBiz struct {
	// 和本机建立的连接
	sessions *xmap.ShardedMap[string, ISession]
}

func NewSessionBiz() SessionBiz {
	return &sessionBiz{
		sessions: xmap.NewShardedMap[string, ISession](64),
	}
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

func (b *sessionBiz) Connect(ctx context.Context, conn ISession) error {
	var (
		uid    = metadata.Uid(ctx)
		device = conn.GetDevice()
	)

	// we try to find some noactive sessions to re-use it
	curSessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "dao failed to get session by uid").WithExtras("uid", uid).WithCtx(ctx)
	}

	var target *dao.Session = pickSessionByStatus(curSessions, ws.StatusTemporayOffline, device, uid)
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
		Reside:         conn.GetLocalIp(),
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

	return nil
}

// disconnect a session
func (b *sessionBiz) Disconnect(ctx context.Context, cid string) error {
	err := infra.Dao().SessionDao.UpdateStatus(ctx, cid, ws.StatusNoActive)
	if err != nil {
		return xerror.Wrapf(err, "dao failed to update status").
			WithExtras("cid", cid).
			WithCtx(ctx)
	}

	b.sessions.Remove(cid)

	return nil
}

// make cids sessions go offline temporarily
func (b *sessionBiz) OfflineAllSessions(ctx context.Context) {
	cids := b.sessions.Keys()

	var wg sync.WaitGroup
	err := slices.BatchAsyncExec(&wg, cids, 200, func(start, end int) error {
		for _, cid := range cids[start:end] {
			err := infra.Dao().SessionDao.UpdateStatus(ctx, cid, ws.StatusTemporayOffline)
			if err != nil {
				xlog.Msgf("session offline dao update status failed").Extras("cid", cid).Err(err).Error()
			}
		}

		return nil
	})

	if err != nil {
		xlog.Msgf("session offline action err").Err(err).Error()
	}

	b.sessions.Clear()
}

func (b *sessionBiz) Heartbeat(ctx context.Context, conn ISession) error {
	// heartbeat
	return b.updateSessionLastActiveTime(ctx, conn)
}

func (b *sessionBiz) updateSessionLastActiveTime(ctx context.Context, conn ISession) error {
	err := infra.Dao().SessionDao.UpdateLastActiveTime(ctx, conn.GetId(), time.Now().Unix())
	if err != nil {
		return xerror.Wrapf(err, "session dao failed to update heartbeat time").
			WithExtras("cid", conn.GetId()).
			WithCtx(ctx)
	}

	return nil
}

func (b *sessionBiz) GetSessionByUid(ctx context.Context, uid int64) ([]ISession, error) {
	logExts := []any{"uid", uid}

	sessions, err := infra.Dao().SessionDao.GetByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "session dao get by uid failed").WithExtras(logExts...).WithCtx(ctx)
	}

	var result []ISession = make([]ISession, 0, len(sessions))
	for _, sess := range sessions {
		if sess.Status == ws.StatusActive {
			result = append(result, b.sessions.Get(sess.Id))
		}
	}

	return result, nil
}

func (b *sessionBiz) GetSessionByUidDevice(ctx context.Context, uid int64, device model.Device) ([]ISession, error) {
	candidates, err := b.GetSessionByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "get session by uid failed").
			WithExtras("uid", uid, "device", device).
			WithCtx(ctx)
	}

	// filter device
	results := make([]ISession, 0, len(candidates))
	for _, c := range candidates {
		if c.GetDevice() == device {
			results = append(results, c)
		}
	}

	return results, nil
}
