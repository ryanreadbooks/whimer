package biz

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra"
	session "github.com/ryanreadbooks/whimer/passport/internal/infra/seesion"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

type CheckInType uint8

const (
	CheckInBySms = 0x3
)

type CheckInRequest struct {
	User *model.UserInfo
	Type CheckInType
	Data struct {
		Platform string
		Sms      string
	}
}

// 登录认证相关
type AccessBiz interface {
	// 验证用户的密码
	VerifyUserPass(ctx context.Context, uid int64, pass string) error
	// 用户登录
	CheckIn(ctx context.Context, req *CheckInRequest) (*model.Session, error)
	// 检查某个sessId 不检查登录的平台
	IsSessIdCheckedIn(ctx context.Context, sessId string) (*model.UserInfo, error)
	// 检查某个sessId是否所属某个平台
	IsSessIdCheckedInPlatform(ctx context.Context, sessId, platform string) (*model.UserInfo, error)
	// 检查用户是否登录
	IsCheckedIn(ctx context.Context, uid int64) (*model.UserInfo, error)
	// 检查某个用户是否在某个平台登录
	IsCheckedInOnPlatform(ctx context.Context, uid int64, platform string) (*model.UserInfo, error)
	// 退出某个登录会话
	CheckOutTarget(ctx context.Context, sessId string) error
	// 全平台退登
	CheckOutAll(ctx context.Context, uid int64) error
}

type accessBiz struct {
	sessMgr *session.Manager
}

func NewAccessBiz() AccessBiz {
	b := &accessBiz{
		sessMgr: session.NewManager(infra.Cache()),
	}

	return b
}

// 验证用户的密码
func (b *accessBiz) VerifyUserPass(ctx context.Context, uid int64, pass string) error {
	passAndSalt, err := infra.Dao().UserDao.FindPassAndSaltByUid(ctx, uid)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return xerror.Wrapf(err, "access biz failed to get pass and salt").WithCtx(ctx)
		}
		// 用户未注册
		return xerror.Wrap(global.ErrUserNotRegister)
	}

	if passAndSalt.Pass != ConfusePassAndSalt(pass, passAndSalt.Salt) {
		return xerror.Wrap(global.ErrPassNotMatch).WithCtx(ctx)
	}

	return nil
}

// 登录
func (b *accessBiz) CheckIn(ctx context.Context, req *CheckInRequest) (*model.Session, error) {
	sess, err := b.sessMgr.NewSession(ctx, req.User.ToUserBase(), req.Data.Platform)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to create new checkin session").WithCtx(ctx)
	}

	return sess, nil
}

func (b *accessBiz) sessionToUserInfo(sess *model.Session) (*model.UserInfo, error) {
	user, err := b.sessMgr.UnmarshalUserBase(sess.Detail)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to unmarshal session")
	}

	return model.NewUserInfoFromUserBase(user), nil
}

func (b *accessBiz) IsSessIdCheckedIn(ctx context.Context, sessId string) (*model.UserInfo, error) {
	sess, err := b.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to examinate session").WithCtx(ctx)
	}

	return b.sessionToUserInfo(sess)
}

func (b *accessBiz) IsSessIdCheckedInPlatform(ctx context.Context, sessId, platform string) (*model.UserInfo, error) {
	sess, err := b.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to examinate session").WithCtx(ctx)
	}

	if sess.Platform != platform {
		return nil, global.ErrSessPlatformNotMatched
	}

	return b.sessionToUserInfo(sess)
}

func (b *accessBiz) CheckOutTarget(ctx context.Context, sessId string) error {
	err := b.sessMgr.InvalidateSession(ctx, sessId)
	if err != nil {
		return xerror.Wrapf(err, "access biz failed to invalidate session").WithCtx(ctx)
	}

	return nil
}

func (b *accessBiz) CheckOutAll(ctx context.Context, uid int64) error {
	err := b.sessMgr.InvalidateAll(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "access biz failed to invalidate all").WithCtx(ctx)
	}
	return nil
}

// 判断用户是否登录
func (b *accessBiz) IsCheckedIn(ctx context.Context, uid int64) (*model.UserInfo, error) {
	sessions, err := b.sessMgr.GetUserSessions(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to get user sessions").WithCtx(ctx)
	}

	if len(sessions) == 0 {
		return nil, global.ErrNotCheckedIn
	}

	return b.sessionToUserInfo(sessions[0])
}

// 检查某个用户是否在某个平台登录
func (b *accessBiz) IsCheckedInOnPlatform(ctx context.Context, uid int64, platform string) (*model.UserInfo, error) {
	sessions, err := b.sessMgr.GetUserSessions(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "access biz failed to get user sessions").WithCtx(ctx)
	}

	if len(sessions) == 0 {
		return nil, global.ErrNotCheckedIn
	}

	var target *model.Session
	for _, sess := range sessions {
		if sess.Platform == platform {
			target = sess
			break
		}
	}

	if target == nil {
		return nil, global.ErrNotCheckedIn
	}

	return b.sessionToUserInfo(target)
}
