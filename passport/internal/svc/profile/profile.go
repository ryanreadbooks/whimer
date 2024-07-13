package profile

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	ptp "github.com/ryanreadbooks/whimer/passport/internal/model/trans/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

	"github.com/zeromicro/go-zero/core/logx"
)

// 获取个人信息
func (s *Service) GetMe(ctx context.Context, uid uint64) (*model.MeInfo, error) {
	basic, err := s.repo.UserBaseRepo.FindBasic(ctx, uid)
	if err != nil {
		logx.Errorf("repo find basic err: %v, uid: %d", err, uid)
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrInternal
		}
		return nil, global.ErrMeNotFound
	}

	return model.NewMeInfoFromUserBasic(basic), nil
}

// 更新个人信息
func (s *Service) UpdateMe(ctx context.Context, newMe *ptp.UpdateMeReq) (*model.MeInfo, error) {
	// 只更新三个字段
	err := s.repo.UserBaseRepo.UpdateBasicCore(ctx, &userbase.Basic{
		Uid:       newMe.Uid,
		Nickname:  newMe.Nickname,
		StyleSign: newMe.StyleSign,
		Gender:    newMe.Gender,
	})

	if err != nil {
		if xsql.IsCriticalErr(err) {
			logx.Errorf("update me err: %v, uid: %d", err, newMe.Uid)
			return nil, global.ErrInternal
		}
		if xsql.IsDuplicate(err) {
			return nil, global.ErrNicknameTaken
		}
		return nil, err
	}

	return &model.MeInfo{
		Uid:       newMe.Uid,
		Nickname:  newMe.Nickname,
		StyleSign: newMe.StyleSign,
		Gender:    model.GenderMap[newMe.Gender],
	}, nil
}

func (s *Service) UpdateTel(ctx context.Context, uid uint64, tel string) error {

	return nil
}

func (s *Service) UpdateEmail(ctx context.Context, uid uint64, email string) error {

	return nil
}

// 上传头像
func (s *Service) UpdateAvatar(ctx context.Context, uid int64, avatar string) error {
	
	return nil
}
