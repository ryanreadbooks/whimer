package access

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/concur"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	tp "github.com/ryanreadbooks/whimer/passport/internal/model/passport"
	"github.com/ryanreadbooks/whimer/passport/internal/model/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

const (
	second      = 1
	minute      = second * 60
	oneMinute   = 1 * minute
	fiveMinutes = 5 * minute
)

// redis key template prefix
const (
	lockKeyForTelSms = "lock:tel:sms:" // lock:tel:sms:%s
	lockKeySignInTel = "lock:signin:tel:"
	keySmsCodeTel    = "sms:code:tel:" // sms:code:tel:%s
)

func getLockTelForSmsKey(tel string) string {
	return lockKeyForTelSms + tel
}

func getSmsCodeTelKey(tel string) string {
	return keySmsCodeTel + tel
}

func getLockSignInTel(tel string) string {
	return lockKeySignInTel + tel
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
func (s *Service) SignInWithSms(ctx context.Context, req *tp.SignInSmdReq) (*userbase.Basic, *model.Session, error) {
	var (
		tel        string = req.Tel
		platform   string = req.Platform
		reqSmsCode string = req.Code
	)

	lock := redis.NewRedisLock(s.cache, getLockSignInTel(tel))
	acquired, err := lock.AcquireCtx(ctx)
	if err != nil {
		logx.Errorf("redis lock fail when sign in with sms, err: %v, tel: %s", err, tel)
		return nil, nil, global.ErrSignIn
	}
	if !acquired {
		return nil, nil, global.ErrSignInTooFrequent
	}
	defer func() {
		if _, err := lock.Release(); err != nil {
			logx.Errorf("redis lock release err: %v", err)
		}
	}()

	smsCode, err := s.cache.GetCtx(ctx, getSmsCodeTelKey(tel))
	if err != nil {
		logx.Errorf("redis get err: %v, tel: %s", err, tel)
		return nil, nil, global.ErrSignIn
	}

	// 检查短信验证码是否正确
	if smsCode != reqSmsCode {
		return nil, nil, global.ErrSmsCodeNotMatch
	}

	// 验证码正确 允许登录
	user, sess, err := s.signIn(ctx, tel, platform)
	if err != nil {
		logx.Errorf("sign in err: %v, tel: %s", err, tel)
		return nil, nil, err
	}

	// 删除验证码
	concur.SafeGo(func() {
		if _, err := s.cache.DelCtx(context.Background(), getSmsCodeTelKey(tel)); err != nil {
			logx.Errorf("cache del sms code err: %v", err)
		}
	})

	return user, sess, nil
}

func (s *Service) signIn(ctx context.Context, tel string, platform string) (*userbase.Basic, *model.Session, error) {
	var (
		user *userbase.Model
		err  error
	)

	// 检查该手机用户是否存在
	user, err = s.repo.UserBaseRepo.FindByTel(ctx, tel)
	if err != nil {
		if !errors.Is(xsql.ErrNoRecord, err) {
			logx.Errorf("user base find basic by tel err: %v, tel: %s", err, tel)
			return nil, nil, global.ErrSignIn
		}

		// 用户未注册 自动注册
		user, err = s.SignUpTel(ctx, tel)
		if err != nil {
			logx.Errorf("auto register tel err: %v, tel: %s", err, tel)
			return nil, nil, global.ErrSignIn
		}
	}

	userBasic := &userbase.Basic{
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

	// 为user登录
	sess, err := s.userSignIn(ctx, userBasic, platform)
	if err != nil {
		logx.Errorf("user sign in err: %v", err)
		return nil, nil, err
	}

	return userBasic, sess, nil
}

func (s *Service) userSignIn(ctx context.Context, user *userbase.Basic, platform string) (*model.Session, error) {
	sess, err := s.sessMgr.NewSession(ctx, user, platform)
	if err != nil {
		logx.Errorf("new session err: %v, uid: %d", err, user.Uid)
		return nil, err
	}

	return sess, nil
}

// 检查是否登录了 不检查登录的平台
func (s *Service) CheckSignedIn(ctx context.Context, sessId string) (*profile.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		logx.Errorf("check signed in get session err: %v, sessId: %s", err, sessId)
		return nil, err
	}

	// 登录了返回用户信息 返回缓存中的用户信息
	return s.ExtractMeInfo(sess.Detail)
}

// 检查对应的平台是否登录了
func (s *Service) CheckPlatformSignedIn(ctx context.Context, sessId string, platform string) (*profile.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		logx.Errorf("check signed in get session err: %v, sessId: %s", err, sessId)
		return nil, err
	}

	if sess.Platform != platform {
		return nil, global.ErrSessPlatformNotMatched
	}

	return s.ExtractMeInfo(sess.Detail)
}

func (s *Service) ExtractMeInfo(detail string) (*profile.MeInfo, error) {
	user, err := s.sessMgr.UnmarshalUserBasic(detail)
	if err != nil {
		logx.Errorf("unmarshal user basic err: %v", err)
		return nil, global.ErrInternal.Msg(err.Error())
	}

	return profile.NewMeInfoFromUserBasic(user), nil
}
