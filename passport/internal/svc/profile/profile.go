package profile

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concur"
	"github.com/ryanreadbooks/whimer/misc/oss/uploader"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/model/profile"
	ptp "github.com/ryanreadbooks/whimer/passport/internal/model/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"
	userrpc "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"

	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Service) replaceAvatar(u *userbase.Basic) {
	if len(u.Avatar) != 0 {
		visit := s.avatarUploader.GetPublicVisitUrl(s.c.Oss.Bucket, u.Avatar, s.c.Oss.DisplayEndpoint)
		u.Avatar = visit
	}
}

// 获取个人信息
func (s *Service) GetMe(ctx context.Context, uid uint64) (*profile.MeInfo, error) {
	basic, err := s.repo.UserBaseRepo.FindBasic(ctx, uid)
	if err != nil {
		xlog.Msg("repo find basic err").Err(err).Extra("target_uid", uid).Errorx(ctx)
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrInternal
		}
		return nil, global.ErrMeNotFound
	}

	s.replaceAvatar(basic)

	return profile.NewMeInfoFromUserBasic(basic), nil
}

// 更新个人信息
func (s *Service) UpdateMe(ctx context.Context, newMe *ptp.UpdateMeReq) (*profile.MeInfo, error) {
	// 只更新三个字段
	err := s.repo.UserBaseRepo.UpdateBasicCore(ctx, &userbase.Basic{
		Uid:       newMe.Uid,
		Nickname:  newMe.Nickname,
		StyleSign: newMe.StyleSign,
		Gender:    newMe.Gender,
	})

	if err != nil {
		if xsql.IsCriticalErr(err) {
			xlog.Msg("update me err").Err(err).Extra("target_uid", newMe.Uid).Errorx(ctx)
			return nil, global.ErrInternal
		}
		if xsql.IsDuplicate(err) {
			return nil, global.ErrNicknameTaken
		}
		return nil, err
	}

	return &profile.MeInfo{
		Uid:       newMe.Uid,
		Nickname:  newMe.Nickname,
		StyleSign: newMe.StyleSign,
		Gender:    profile.GenderMap[newMe.Gender],
	}, nil
}

func (s *Service) UpdateTel(ctx context.Context, tel string) error {

	return nil
}

func (s *Service) UpdateEmail(ctx context.Context, email string) error {

	return nil
}

// 上传头像
func (s *Service) UpdateAvatar(ctx context.Context, req *profile.AvatarInfoReq) (string, error) {
	var (
		user    = model.CtxGetMeInfo(ctx)
		objKey  = s.avatarKeyGen.Gen()
		objName = objKey + req.Ext
	)

	// content上传到oss
	err := s.avatarUploader.Upload(ctx, &uploader.UploadMeta{
		Bucket:      s.c.Oss.Bucket,
		Name:        objName,
		Buf:         req.Content,
		ContentType: req.ContentType,
	})
	if err != nil {
		xlog.Msg("avatar upload err").Err(err).Extra("target_uid", user.Uid).Errorx(ctx)
		return "", global.ErrUploadAvatar
	}

	// avatar数据落库
	err = s.repo.UserBaseRepo.UpdateAvatar(ctx, objName, user.Uid)
	if err != nil {
		xlog.Msg("repo update avatar err").Err(err).Extra("target_uid", user.Uid).Errorx(ctx)
		// 异步
		concur.DoneIn(time.Second*10, func(ctx context.Context) {
			if err := s.avatarUploader.Remove(ctx, s.c.Oss.Bucket, objName); err != nil {
				xlog.Msg("goroutine repo update then remove oss err").Err(err).
					Extra("target_uid", user.Uid).
					Extra("obj", objName).
					Errorx(ctx)
			}
		})

		return "", global.ErrUploadAvatar
	}

	// 返回头像访问链接
	visitUrl := s.avatarUploader.GetPublicVisitUrl(s.c.Oss.Bucket, objName, s.c.Oss.DisplayEndpoint)

	return visitUrl, nil
}

// 批量获取uid的信息
func (s *Service) GetByUids(ctx context.Context, uids []uint64) (map[string]*userrpc.UserInfo, error) {
	basics, err := s.repo.UserBaseRepo.FindBasicByUids(ctx, uids)
	if err != nil {
		logx.Errorw("repo find basic by uids err", xlog.WithUid(ctx), logx.Field("uids", uids))
		return nil, global.ErrGetUserFail
	}

	resp := make(map[string]*userrpc.UserInfo, len(basics))
	for _, info := range basics {
		s.replaceAvatar(info)
		resp[strconv.FormatUint(info.Uid, 10)] = basicToUserInfoPb(info)
	}

	return resp, nil
}

func basicToUserInfoPb(b *userbase.Basic) *userrpc.UserInfo {
	return &userrpc.UserInfo{
		Uid:       b.Uid,
		Nickname:  b.Nickname,
		StyleSign: b.StyleSign,
		Avatar:    b.Avatar,
		Gender:    profile.GenderMap[b.Gender],
	}
}
