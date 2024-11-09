package biz

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/infra"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/passport/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
)

const (
	idgenRegIdKey = "passport:uid:id:w"
	idgenStep     = 10000
)

// 用户注册
type RegisterBiz interface {
	// 新用户注册通过手机号注册
	UserRegister(ctx context.Context, tel string) (*model.UserInfo, error)
}

type registerBiz struct {
}

func NewRegisterBiz() RegisterBiz {
	b := &registerBiz{}

	return b
}

func (b *registerBiz) UserRegister(ctx context.Context, tel string) (*model.UserInfo, error) {
	var (
		now = time.Now().Unix()
	)

	// 分配用户id
	uid, err := dep.IdGen().GetId(ctx, idgenRegIdKey, idgenStep)
	if err != nil {
		return nil, xerror.Wrapf(global.ErrRegisterTel, "register biz failed to get new uid").WithExtra("cause", err)
	}

	// 随机生成初始密码
	pass, salt, err := MakeRandomInitPass()
	if err != nil {
		return nil, xerror.Wrapf(global.ErrRegisterTel, "register biz failed to init random password").WithExtra("cause", err)
	}

	// 随机用户名
	nickname := makeInitNickname(uid)
	data := &dao.User{
		UserBase: dao.UserBase{
			Uid:      uid,
			Nickname: nickname,
			Avatar:   "", // 默认头像在各端处理
			Tel:      tel,
			Timing: dao.Timing{
				CreateAt: now,
				UpdateAt: now,
			},
		},
		UserSecret: dao.UserSecret{
			Pass: pass,
			Salt: salt,
		},
	}

	err = infra.Dao().UserDao.Insert(ctx, data)
	if err != nil {
		if errors.Is(err, xsql.ErrDuplicate) {
			return nil, global.ErrTelTaken
		}
		return nil, xerror.Wrapf(err, "register biz failed to insert new user")
	}

	return model.NewUserInfoFromUserBase(&data.UserBase), nil
}

func makeInitNickname(uid uint64) string {
	return fmt.Sprintf("whimer_%d", uid)
}
