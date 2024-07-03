package login

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
	"github.com/ryanreadbooks/whimer/passport/internal/repo/userbase"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/logx"
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

// 生成随机盐
func makeSalt() (string, error) {
	istn, err := uuid.NewUUID()
	if err != nil {
		return "", err
	}

	suffix := utils.RandomString(8)
	cand := istn.String() + suffix

	hasher := sha256.New()
	hasher.Write([]byte(cand))
	salt := hasher.Sum(nil)

	return hex.EncodeToString(salt), nil
}

func concatPassSalt(pass, salt string) string {
	idx := len(salt) / 2

	concat := salt[0:idx] + pass + salt[idx:]

	sh := sha256.New()
	sh.Write([]byte(concat))
	res := sh.Sum(nil)

	return hex.EncodeToString(res)
}

// 生成随机初始密码 与其对应的盐
func makeInitPass() (pass, salt string, err error) {
	salt, err = makeSalt()
	if err != nil {
		return
	}

	// 生成随机密码
	rawPass := utils.RandomPass(10) // 随机生成10位密码
	// 随机密码拼接盐
	pass = concatPassSalt(rawPass, salt)

	return
}

// 初始化新注册用户的昵称
func makeInitNickname(uid uint64) string {
	return fmt.Sprintf("野生的用户_%d", uid)
}

func (s *Service) regTakeUid(ctx context.Context) (uint64, error) {
	return 0, nil
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
