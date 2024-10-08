package access

import (
	"context"
	"errors"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/concur"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	tp "github.com/ryanreadbooks/whimer/passport/internal/model/passport"
	"github.com/ryanreadbooks/whimer/passport/internal/model/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

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
		xlog.Msg("redis lock fail when request sms err").Err(err).Errorx(ctx)
		return global.ErrRequestSms
	}

	if !acquired {
		return global.ErrRequestSmsFrequent
	}

	smsCode := makeSmsCode()
	key := getSmsCodeTelKey(tel)
	err = s.cache.SetexCtx(ctx, key, smsCode, fiveMinutes)
	if err != nil {
		xlog.Msg("request sms redix setex err").Err(err).Errorx(ctx)
		return global.ErrRequestSms
	}

	xlog.Msg(fmt.Sprintf("request sms, code = %s", smsCode))

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
		xlog.Msg("redis lock fail when request sms err").Err(err).Errorx(ctx)
		return nil, nil, global.ErrSignIn
	}
	if !acquired {
		return nil, nil, global.ErrSignInTooFrequent
	}
	defer func() {
		if _, err := lock.Release(); err != nil {
			xlog.Msg("redis lock release err").Err(err).Errorx(ctx)
		}
	}()

	smsCode, err := s.cache.GetCtx(ctx, getSmsCodeTelKey(tel))
	if err != nil {
		xlog.Msg("redis get err").Err(err).Errorx(ctx)
		return nil, nil, global.ErrSignIn
	}

	// 检查短信验证码是否正确
	if smsCode != reqSmsCode {
		return nil, nil, global.ErrSmsCodeNotMatch
	}

	// 验证码正确 允许登录
	user, sess, err := s.signIn(ctx, tel, platform)
	if err != nil {
		xlog.Msg("sign in err").Err(err).Errorx(ctx)
		return nil, nil, err
	}

	// 删除验证码
	concur.SafeGo(func() {
		ctxc := context.WithoutCancel(ctx)
		if _, err := s.cache.DelCtx(ctxc, getSmsCodeTelKey(tel)); err != nil {
			xlog.Msg("cache del sms code err").Err(err).Errorx(ctxc)
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
			xlog.Msg("user base find basic by tel err").Err(err).Errorx(ctx)
			return nil, nil, global.ErrSignIn
		}

		// 用户未注册 自动注册
		user, err = s.SignUpTel(ctx, tel)
		if err != nil {
			xlog.Msg("auto register tel err").Err(err).Errorx(ctx)
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
		xlog.Msg("user sign in err").Err(err).Errorx(ctx)
		return nil, nil, err
	}

	return userBasic, sess, nil
}

func (s *Service) userSignIn(ctx context.Context, user *userbase.Basic, platform string) (*model.Session, error) {
	sess, err := s.sessMgr.NewSession(ctx, user, platform)
	if err != nil {
		xlog.Msg("new session err").Err(err).Uid(user.Uid).Error()
		return nil, err
	}

	return sess, nil
}

// 检查是否登录了 不检查登录的平台
func (s *Service) CheckSignedIn(ctx context.Context, sessId string) (*profile.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		xlog.Msg("check signed in get session err").Err(err).Extra("sessId", sessId).Error()
		return nil, err
	}

	// 登录了返回用户信息 返回缓存中的用户信息
	return s.ExtractMeInfo(sess.Detail)
}

// 检查对应的平台是否登录了
func (s *Service) CheckPlatformSignedIn(ctx context.Context, sessId string, platform string) (*profile.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		xlog.Msg("check signed in get session err").Err(err).Extra("sessId", sessId).Error()
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
		xlog.Msg("unmarshal user basic err").Err(err).Error()
		return nil, global.ErrInternal.Msg(err.Error())
	}

	return profile.NewMeInfoFromUserBasic(user), nil
}

func (s *Service) SignOutCurrent(ctx context.Context, sessId string) error {
	err := s.sessMgr.InvalidateSession(ctx, sessId)
	if err != nil {
		xlog.Msg("signout single invalidate session err").Err(err).Extra("sessId", sessId).Error()
		return global.ErrInternal
	}

	return nil
}
