package access

import (
	"context"
	"errors"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"

	"github.com/zeromicro/go-zero/core/logx"
)

func (s *Service) Me(ctx context.Context, sessId string) (*model.MeInfo, error) {
	sess, err := s.sessMgr.GetSession(ctx, sessId)
	if err != nil {
		if !errors.Is(err, global.ErrSessInvalidated) {
			logx.Errorf("get session err: %v, sessId: %s", err, sessId)
		}
		return nil, err
	}

	user, err := s.sessMgr.UnmarshalUserBasic(sess.Detail)
	if err != nil {
		logx.Errorf("unmarshal user basic err: %v, sessId: %s", err, sessId)
		return nil, global.ErrInternal.Msg(err.Error())
	}

	return model.NewMeInfoFromUserBasic(user), nil
}
