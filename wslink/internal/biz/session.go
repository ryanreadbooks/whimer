package biz

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/utils/slices"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
)

type SessionBiz interface {
	Connect(ctx context.Context, conn *ws.Connection) error
	Disconnect(ctx context.Context, cid string) error
	OfflineSession(ctx context.Context, cids []string)
}

type sessionBiz struct{}

func NewSessionBiz() SessionBiz {
	return &sessionBiz{}
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

func (b *sessionBiz) Connect(ctx context.Context, conn *ws.Connection) error {
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
		Reside:         config.GetIpAndPort(),
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

	return nil
}

func (b *sessionBiz) OfflineSession(ctx context.Context, cids []string) {
	// make cids sessions go offline temporarily
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
}
