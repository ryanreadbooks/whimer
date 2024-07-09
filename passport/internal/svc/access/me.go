package access

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Service) extractMeInfo(detail string) (*model.MeInfo, error) {
	user, err := s.sessMgr.UnmarshalUserBasic(detail)
	if err != nil {
		logx.Errorf("unmarshal user basic err: %v", err)
		return nil, global.ErrInternal.Msg(err.Error())
	}

	return model.NewMeInfoFromUserBasic(user), nil
}

func (s *Service) Me(ctx context.Context, sessId string) (*model.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		if !errors.Is(err, global.ErrSessInvalidated) {
			logx.Errorf("get session err: %v, sessId: %s", err, sessId)
		}
		return nil, err
	}

	// 去数据库取最新的个人信息
	basic, err := s.repo.UserBaseRepo.FindBasic(ctx, sess.Uid)
	if err != nil {
		logx.Errorf("repo find basic err: %v, uid: %d", err, sess.Uid)
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrInternal
		}
		return nil, global.ErrMeNotFound
	}

	return model.NewMeInfoFromUserBasic(basic), nil
}
