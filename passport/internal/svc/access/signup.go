package access

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

)

const (
	foliumRegIdKey = "passport:uid:id:w"
)

// 初始化新注册用户的昵称
func makeInitNickname(uid uint64) string {
	return fmt.Sprintf("whimer_%d", uid)
}

func (s *Service) regTakeUid(ctx context.Context) (uint64, error) {
	return s.idgen.GetId(ctx, foliumRegIdKey, 10000)
}

// 通过电话注册账号
func (s *Service) SignUpTel(ctx context.Context, tel string) (*userbase.Model, error) {
	var (
		now = time.Now().Unix()
	)

	// 1. 初始化密码
	pass, salt, err := makeInitPass()
	if err != nil {
		xlog.Msg("gen init pass when register tel err").Err(err).Error()
		return nil, global.ErrRegisterTel
	}

	// 2. 生成uid
	uid, err := s.regTakeUid(ctx)
	if err != nil {
		xlog.Msg("reg take uid err").Err(err).Error()
		return nil, global.ErrRegisterTel
	}

	// 3. 生成随机昵称
	nickname := makeInitNickname(uid)

	user := userbase.Model{
		Uid:      uid,
		Nickname: nickname,
		Avatar:   "", // 默认头像在各端处理
		Tel:      tel,
		Pass:     pass,
		Salt:     salt,
		Timing: userbase.Timing{
			CreateAt: now,
			UpdateAt: now,
		},
	}

	err = s.repo.UserBaseRepo.Insert(ctx, &user)
	if err != nil {
		xlog.Msg("register tel insert user err").Err(err).Error()
		if errors.Is(xsql.ErrDuplicate, err) {
			// 手机号重复
			return nil, global.ErrTelTaken
		}
		return nil, global.ErrRegisterTel
	}

	return &user, nil
}
