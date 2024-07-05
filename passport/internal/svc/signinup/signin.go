package signinup

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
	tp "github.com/ryanreadbooks/whimer/passport/internal/types/passport"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	second      = 1
	minute      = second * 60
	oneMinute   = 1 * minute
	fiveMinutes = 5 * minute
)

// redis key template
const (
	lockTelForSmsKey = "lock:tel:sms:" // lock:tel:sms:%s
	smsCodeTelKey    = "sms:code:tel:" // sms:code:tel:%s
	lockSignInTel    = "lock:signin:tel:"
)

func getLockTelForSmsKey(tel string) string {
	return lockTelForSmsKey + tel
}

func getSmsCodeTelKey(tel string) string {
	return smsCodeTelKey + tel
}

func getLockSignInTel(tel string) string {
	return lockSignInTel + tel
}

// 请求发送手机验证码
func (s *Service) RequestSms(ctx context.Context, tel string) error {
	lock := redis.NewRedisLock(s.cache, getLockTelForSmsKey(tel))
	lock.SetExpire(minute) // 同一个电话号码60s只能获取一次手机验证码
	acquired, err := lock.AcquireCtx(ctx)
	if err != nil {
		logx.Errorf("redis lock fail when request sms, err: %v, tel: %s", err, tel)
		return global.ErrRequestSms
	}

	if !acquired {
		return global.ErrRequestSmsFrequent
	}

	smsCode := makeSmsCode()
	key := getSmsCodeTelKey(tel)
	err = s.cache.SetexCtx(ctx, key, smsCode, fiveMinutes)
	if err != nil {
		logx.Errorf("request sms redix setex err: %v, tel: %s", err, tel)
		return global.ErrRequestSms
	}

	logx.Debugf("request sms, code = %s", smsCode)

	return nil
}

// 短信验证码登录
func (s *Service) SignInWithSms(ctx context.Context, req *tp.SignInSmdReq) (*userbase.Basic, error) {
	var (
		tel        string = req.Tel
		reqSmsCode string = req.Code
	)

	lock := redis.NewRedisLock(s.cache, getLockSignInTel(tel))
	acquired, err := lock.AcquireCtx(ctx)
	if err != nil {
		logx.Errorf("redis lock fail when sign in with sms, err: %v, tel: %s", err, tel)
		return nil, global.ErrSignIn
	}
	if !acquired {
		return nil, global.ErrSignInTooFrequent
	}
	defer func() {
		if _, err := lock.Release(); err != nil {
			logx.Errorf("redis lock release err: %v", err)
		}
	}()

	smsCode, err := s.cache.GetCtx(ctx, getSmsCodeTelKey(tel))
	if err != nil {
		logx.Errorf("redis get err: %v, tel: %s", err, tel)
		return nil, global.ErrSignIn
	}

	if smsCode != reqSmsCode {
		return nil, global.ErrSmsCodeNotMatch
	}

	// 验证码正确 允许登录
	user, err := s.signIn(ctx, tel)
	if err != nil {
		logx.Errorf("sign in err: %v, tel: %s", err, tel)
		return nil, err
	}

	return user, nil
}

func (s *Service) signIn(ctx context.Context, tel string) (*userbase.Basic, error) {
	// 检查该手机用户是否存在
	var (
		user *userbase.Model
		err  error
	)

	user, err = s.repo.UserBaseRepo.FindByTel(ctx, tel)
	if err != nil {
		if !errors.Is(xsql.ErrNoRecord, err) {
			logx.Errorf("user base find basic by tel err: %v, tel: %s", err, tel)
			return nil, global.ErrSignIn
		}

		// 用户未注册 自动注册
		user, err = s.SignUpTel(ctx, tel)
		if err != nil {
			logx.Errorf("auto register tel err: %v, tel: %s", err, tel)
			return nil, global.ErrSignIn
		}
	}

	// TODO 验证user是否已经登录 为user登录

	basic := &userbase.Basic{
		Uid:       user.Uid,
		Nickname:  user.Nickname,
		Avatar:    user.Avatar,
		StyleSign: user.StyleSign,
		Gender:    user.Gender,
		Timing: userbase.Timing{
			CreateAt: user.CreateAt,
			UpdateAt: user.UpdateAt,
		},
	}

	return basic, nil
}
