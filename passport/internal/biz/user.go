package biz

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/imgproxy"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/oss/uploader"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

// 用户个人信息相关
// 包含通用功能
type UserBiz interface {
	// 获取个人信息
	GetUser(ctx context.Context, uid int64) (*model.UserInfo, error)
	// 批量和获取用户信息
	BatchGetUser(ctx context.Context, uids []int64) (map[int64]*model.UserInfo, error)
	// 更新个人信息
	UpdateUser(ctx context.Context, req *model.UpdateUserRequest) (*model.UserInfo, error)
	// 上传头像
	UpdateAvatar(ctx context.Context, uid int64, req *model.AvatarInfoRequest) (string, error)
	// 通过手机号获取用户
	GetUserByTel(ctx context.Context, tel string) (*model.UserInfo, error)
	// 获取头像链接
	ReplaceAvatar(u *model.UserInfo)
	ReplaceAvatarUrl(url string) string
}

type userBiz struct {
	avatarKeyGen   *keygen.Generator
	avatarUploader *uploader.Uploader
}

func NewUserBiz() UserBiz {
	b := &userBiz{
		avatarKeyGen: keygen.NewGenerator(
			keygen.WithBucket(config.Conf.Oss.Bucket),
			keygen.WithPrefix(config.Conf.Oss.Prefix),
			keygen.WithPrependPrefix(true),
			keygen.WithPrependBucket(false), // 生成key的时候不需要附带上bucket
		),
	}

	avatartUploader, err := uploader.New(uploader.Config{
		Ak:       config.Conf.Oss.Ak,
		Sk:       config.Conf.Oss.Sk,
		Endpoint: config.Conf.Oss.Endpoint,
		Location: config.Conf.Oss.Location,
	})
	if err != nil {
		panic(err)
	}

	b.avatarUploader = avatartUploader

	return b
}

func (b *userBiz) getUser(ctx context.Context, uid int64) (*model.UserInfo, error) {
	user, err := infra.Dao().UserDao.FindUserBaseByUid(ctx, uid)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrapf(err, "user biz failed to find user").WithExtra("uid", uid).WithCtx(ctx)
		}
		return nil, global.ErrUserNotFound
	}

	return model.NewUserInfoFromUserBase(user), nil
}

func (b *userBiz) GetUser(ctx context.Context, uid int64) (*model.UserInfo, error) {
	user, err := b.getUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	b.ReplaceAvatar(user)

	return user, nil
}

// 更新个人信息
// 只允许更新昵称，个性签名和性别
func (b *userBiz) UpdateUser(ctx context.Context, req *model.UpdateUserRequest) (*model.UserInfo, error) {
	user, err := b.getUser(ctx, req.Uid)
	if err != nil {
		return nil, err
	}

	err = infra.Dao().UserDao.UpdateUserBase(ctx, &dao.UserBase{
		Uid:       user.Uid,
		Nickname:  req.Nickname,
		StyleSign: req.StyleSign,
		Gender:    req.Gender,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "user biz failed to update user base").WithExtra("req", req).WithCtx(ctx)
	}

	return &model.UserInfo{
		Uid:       req.Uid,
		Nickname:  req.Nickname,
		StyleSign: req.StyleSign,
		Gender:    model.GenderMap[req.Gender],
	}, nil
}

// 上传新头像
func (b *userBiz) UpdateAvatar(ctx context.Context, uid int64, req *model.AvatarInfoRequest) (string, error) {
	var (
		objKey  = b.avatarKeyGen.Gen()
		objName = objKey + req.Ext
		bucket  = config.Conf.Oss.Bucket
	)

	// content上传到oss
	err := b.avatarUploader.Upload(ctx, &uploader.UploadMeta{
		Bucket:      bucket,
		Name:        objName,
		Buf:         req.Content,
		ContentType: req.ContentType,
	})
	if err != nil {
		return "", xerror.Wrapf(global.ErrUploadAvatar, "user biz failed to upload avatar to oss").
			WithExtras("bucket", bucket, "objName", objName, "cause", err).
			WithCtx(ctx)
	}

	// avatar key落库
	err = infra.Dao().UserDao.UpdateAvatar(ctx, objName, uid)
	if err != nil {
		// 尽可能删除成功上传的avatar
		concurrent.DoneIn(time.Second*25, func(gctx context.Context) {
			if err2 := b.avatarUploader.Remove(gctx, bucket, objName); err2 != nil {
				xlog.Msg("user biz remove oss failed").Err(err2).
					Extra("targetUid", uid).
					Extra("bucket", bucket).
					Extra("obj", objName).
					Errorx(ctx)
			}
		})

		return "", xerror.Wrapf(err, "user biz failed to update avatar").WithExtras("objKey", objKey)
	}

	// 返回头像访问链接
	visitUrl := getVisitUrlFromConf(objName)

	return visitUrl, nil
}

func getVisitUrlFromConf(avtr string) string {
	visitUrl := imgproxy.GetSignedUrl(
		config.Conf.Oss.AvatarDisplayEndpoint(), // 带上桶的名称,不加前缀/
		config.Conf.Oss.Bucket+"/"+avtr,
		config.Conf.ImgProxyAuth.GetKey(),
		config.Conf.ImgProxyAuth.GetSalt(),
	)
	return visitUrl
}

func (b *userBiz) ReplaceAvatar(u *model.UserInfo) {
	if len(u.Avatar) > 0 {
		u.Avatar = getVisitUrlFromConf(u.Avatar)
	}
}

func (b *userBiz) ReplaceAvatarUrl(url string) string {
	if len(url) > 0 {
		return getVisitUrlFromConf(url)
	}

	return url
}

func (b *userBiz) BatchGetUser(ctx context.Context, uids []int64) (map[int64]*model.UserInfo, error) {
	users, err := infra.Dao().UserDao.FindUserBaseByUids(ctx, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "user biz failed to get users").WithExtra("uids", uids)
	}

	resp := make(map[int64]*model.UserInfo, len(users))
	for _, user := range users {
		info := model.NewUserInfoFromUserBase(user)
		b.ReplaceAvatar(info)
		resp[info.Uid] = info
	}

	return resp, nil
}

func (b *userBiz) GetUserByTel(ctx context.Context, tel string) (*model.UserInfo, error) {
	user, err := infra.Dao().UserDao.FindUserBaseByTel(ctx, tel)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrapf(err, "user biz failed to find user by tel").WithCtx(ctx)
		}
		return nil, global.ErrUserNotFound
	}

	return model.NewUserInfoFromUserBase(user), nil
}
