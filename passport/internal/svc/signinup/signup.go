package signinup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

	"github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	foliumRegIdKey = "passport:uid:id:w"
)

type Service struct {
	c    *config.Config
	repo *repo.Repo

	idgen *sdk.Client
}

func New(c *config.Config, repo *repo.Repo) *Service {
	s := &Service{
		c:    c,
		repo: repo,
	}

	var err error
	s.idgen, err = sdk.NewClient(sdk.WithGrpc(s.c.Idgen.Addr))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = s.idgen.Ping(ctx)
	if err != nil {
		logx.Errorf("new passport svc, can not ping idgen(folium): %v", err)
	}

	return s
}



// 初始化新注册用户的昵称
func makeInitNickname(uid uint64) string {
	return fmt.Sprintf("野生的用户_%d", uid)
}

func (s *Service) regTakeUid(ctx context.Context) (uint64, error) {
	return s.idgen.GetId(ctx, foliumRegIdKey, 10000)
}

// 通过电话注册账号
func (s *Service) RegisterTel(ctx context.Context, tel string) (*userbase.Model, error) {
	// 1. 初始化密码
	pass, salt, err := makeInitPass()
	if err != nil {
		logx.Errorf("gen init pass when register tel err: %v", err)
		return nil, global.ErrRegisterTel
	}

	// 2. 生成uid
	uid, err := s.regTakeUid(ctx)
	if err != nil {
		logx.Errorf("reg take uid err: %v", err)
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
	}

	err = s.repo.UserBaseRepo.Insert(ctx, &user)
	if err != nil {
		logx.Errorf("register tel insert user err: %v", err)
		if errors.Is(xsql.ErrDuplicate, err) {
			// 手机号重复
			return nil, global.ErrTelTaken
		}
		return nil, global.ErrRegisterTel
	}

	return &user, nil
}
