package biz

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dep"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type AccessSmsBiz interface {
	// 发送短信
	RequestSendSms(ctx context.Context, tel string) error
	// 验证码是否正确
	CheckSmsCorrect(ctx context.Context, tel, smsCode string) error
	// 删除验证码
	DeleteSmsCode(ctx context.Context, tel string) error
}

type accessSmsBiz struct{}

func NewAccessSmsBiz() AccessSmsBiz {
	b := &accessSmsBiz{}

	return b
}

// redis key template prefix
const (
	lockKeyForTelSms  = "lock:tel:sms:" // lock:tel:sms:%s
	keySmsCodeTel     = "sms:code:tel:" // sms:code:tel:%s
	lockKeyCheckInTel = "lock:checkin:tel:"
)

const (
	second      = 1
	minute      = second * 60
	oneMinute   = 1 * minute
	fiveMinutes = 5 * minute
)

const (
	smsTemplate = "【whimer】您正在通过短信登录whimer，当前验证码为%s，该验证码5分钟内有效，请勿泄露于他人。"
)

func getLockTelForSmsKey(tel string) string {
	return lockKeyForTelSms + tel
}

func getSmsCodeTelKey(tel string) string {
	return keySmsCodeTel + tel
}

func getLockCheckInTel(tel string) string {
	return lockKeyCheckInTel + tel
}

// 发送验证码
func (b *accessSmsBiz) RequestSendSms(ctx context.Context, tel string) (err error) {
	rawTel := tel
	// tel加密存储
	tel, err = infra.Encryptor().Encrypt(ctx, tel)
	if err != nil {
		return xerror.Wrapf(err, "sms biz failed to encrypt tel")
	}

	lock := redis.NewRedisLock(infra.Cache(), getLockTelForSmsKey(tel))
	lock.SetExpire(minute) // 同一个电话号码60s只能获取一次验证码
	acquired, err := lock.AcquireCtx(ctx)
	if err != nil {
		return xerror.Wrapf(global.ErrRequestSms, "access sms biz failed to acquire lock when sending sms").WithExtra("cause", err).WithCtx(ctx)
	}

	if !acquired {
		return global.ErrRequestSmsFrequent
	}

	smsCode := MakeSmsCode()
	if config.Conf.IsDev { // 开发环境下固定smscode
		if config.Conf.DevSmsCode != "" {
			smsCode = config.Conf.DevSmsCode
		}
	}
	err = infra.Cache().SetexCtx(ctx, getSmsCodeTelKey(tel), smsCode, 5*minute)
	if err != nil {
		return xerror.Wrapf(global.ErrRequestSms, "access sms biz failed to set smscode in cache").WithExtra("cause", err).WithCtx(ctx)
	}

	err = dep.SmsSender().Send(ctx, rawTel, fmt.Sprintf(smsTemplate, smsCode))
	if err != nil {
		return xerror.Wrapf(global.ErrRequestSms, "access sms biz failed to send sms").WithExtra("cause", err).WithCtx(ctx)
	}

	return nil
}

// 手机号验证码登录
func (b *accessSmsBiz) CheckSmsCorrect(ctx context.Context, tel, smsCode string) (err error) {
	tel, err = encryptTel(ctx, tel)
	if err != nil {
		return err
	}

	lock := redis.NewRedisLock(infra.Cache(), getLockCheckInTel(tel))
	acquired, err := lock.AcquireCtx(ctx)
	if err != nil {
		err = xerror.Wrapf(global.ErrRequestSms, "access sms biz failed to acquire lock when checking in").WithExtra("cause", err).WithCtx(ctx)
		return
	}
	defer func() {
		if _, err2 := lock.ReleaseCtx(ctx); err2 != nil {
			xlog.Msg("access sms biz failed to release lock").Err(err2).Errorx(ctx)
		}
	}()

	if !acquired {
		err = global.ErrCheckInTooFrequent
		return
	}

	target, err := infra.Cache().GetCtx(ctx, getSmsCodeTelKey(tel))
	if err != nil {
		err = xerror.Wrapf(global.ErrCheckIn, "access sms biz failed to get smscode from cache").WithExtra("cause", err).WithCtx(ctx)
		return
	}

	// 检查验证码是否正确
	if smsCode != target {
		err = global.ErrSmsCodeNotMatch
		return
	}

	return
}

func (b *accessSmsBiz) DeleteSmsCode(ctx context.Context, tel string) (err error) {
	tel, err = encryptTel(ctx, tel)
	if err != nil {
		return err
	}

	if _, err := infra.Cache().DelCtx(ctx, getSmsCodeTelKey(tel)); err != nil {
		return xerror.Wrapf(global.ErrAccessBiz, "access sms biz failed to del smscode from cache").WithExtra("cause", err).WithCtx(ctx)
	}

	return nil
}

func encryptTel(ctx context.Context, tel string) (string, error) {
	tel, err := infra.Encryptor().Encrypt(ctx, tel)
	if err != nil {
		return "", xerror.Wrapf(err, "sms biz failed to encrypt tel")
	}

	return tel, nil
}
