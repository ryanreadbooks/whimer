package srv

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/passport/internal/biz"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

type AccessSrv struct {
	parent *Service

	accessBiz   biz.AccessBiz
	smsBiz      biz.AccessSmsBiz
	userBiz     biz.UserBiz
	registerBiz biz.RegisterBiz
}

func NewAccessSrv(p *Service, biz biz.Biz) *AccessSrv {
	return &AccessSrv{
		parent:      p,
		accessBiz:   biz.Access,
		smsBiz:      biz.AccessSms,
		userBiz:     biz.User,
		registerBiz: biz.Register,
	}
}

// 发送验证码
func (s *AccessSrv) SendSmsCode(ctx context.Context, tel string) error {
	return s.smsBiz.RequestSendSms(ctx, tel)
}

// 手机号+短信验证码登录
func (s *AccessSrv) SmsCheckIn(ctx context.Context, req *model.SmsCheckInRequest) (resp *model.CheckInResponse, err error) {
	var (
		tel      = req.Tel // not encrypted
		platform = req.Platform
		smsCode  = req.Code
	)

	// 检查该手机用户是否存在
	user, err := s.userBiz.GetUserByTel(ctx, tel)
	if err != nil && !errors.Is(err, global.ErrUserNotFound) {
		err = xerror.Wrapf(err, "access srv failed to get user by tel").WithCtx(ctx)
		return
	}

	// 检查验证码是否正确
	errCodeCheck := s.smsBiz.CheckSmsCorrect(ctx, tel, smsCode)
	if errCodeCheck != nil {
		err = xerror.Wrapf(errCodeCheck, "access srv check sms correctness failed").WithCtx(ctx)
		return
	}

	// 用户不存在自动注册
	if errors.Is(err, global.ErrUserNotFound) {
		user, err = s.registerBiz.UserRegister(ctx, tel)
		if err != nil {
			err = xerror.Wrapf(err, "access srv failed to auto register for user").WithCtx(ctx)
			return
		}
	}

	checkinReq := biz.CheckInRequest{
		User: user,
		Type: biz.CheckInBySms,
	}
	checkinReq.Data.Sms = smsCode
	checkinReq.Data.Platform = platform

	// 为user登录
	session, err := s.accessBiz.CheckIn(ctx, &checkinReq)
	if err != nil {
		err = xerror.Wrapf(err, "access srv failed to checkin").WithCtx(ctx)
		return
	}

	resp = model.NewCheckInResponseFromUserInfo(user)
	resp.Session = session

	err = s.smsBiz.DeleteSmsCode(ctx, tel)
	if err != nil {
		xlog.Msg("service delete smscode failed").Err(err).Errorx(ctx)
	}

	// TODO 将当前平台中已经生效的会话清除

	resp.Avatar = s.userBiz.ReplaceAvatarUrl(resp.Avatar)

	return resp, nil
}

func (s *AccessSrv) IsCheckedIn(ctx context.Context, sessId string) (*model.UserInfo, error) {
	return s.accessBiz.IsSessIdCheckedIn(ctx, sessId)
}

func (s *AccessSrv) IsCheckedInOnPlatform(ctx context.Context, sessId, plat string) (*model.UserInfo, error) {
	return s.accessBiz.IsSessIdCheckedInPlatform(ctx, sessId, plat)
}

// TODO 允许使用密码登录
func (s *AccessSrv) PassCheckIn(ctx context.Context, req *model.PassCheckInRequest) error {
	return nil
}

// 登出
func (s *AccessSrv) CheckOutCurrent(ctx context.Context, sessId string) error {
	return s.accessBiz.CheckOutTarget(ctx, sessId)
}

func (s *AccessSrv) CheckoutAll(ctx context.Context) error {
	var (
		userInfo = model.CtxGetUserInfo(ctx)
		uid      = userInfo.Uid
	)

	if uid == 0 {
		return global.ErrNotCheckedIn
	}

	return s.accessBiz.CheckOutAll(ctx, uid)
}
