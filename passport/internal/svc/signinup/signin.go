package signinup

import (
	"context"
	"fmt"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
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
	lockTelForSmsKey = "lock:tel:sms:%s"
	smsCodeTelKey    = "sms:code:tel:%s"
)

// 请求发送手机验证码
func (s *Service) RequestSms(ctx context.Context, tel string) error {
	lock := redis.NewRedisLock(s.cache, fmt.Sprintf(lockTelForSmsKey, tel))
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
	key := fmt.Sprintf(smsCodeTelKey, tel)
	err = s.cache.SetexCtx(ctx, key, smsCode, fiveMinutes)
	if err != nil {
		logx.Errorf("request sms redix setex err: %v, tel: %s", err, tel)
		return global.ErrRequestSms
	}

	logx.Debugf("request sms, code = %s", smsCode)

	return nil
}

// 短信验证码登录
func (s *Service) SignInWithSms(ctx context.Context) error {

	return nil
}
