package access

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"

	"github.com/google/uuid"
)

func makeSmsCode() string {
	// 0-899999 + 100000 = 100000 ~ 999999
	return fmt.Sprintf("%06d", rand.Int31n(900000)+100000)
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

// 验证用户的密码是否正确
func (s *Service) verifyPass(ctx context.Context, uid uint64, pass string) error {
	passSalt, err := s.repo.UserBaseRepo.FindPassSalt(ctx, uid)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			xlog.Msg("pass verification err").Err(err).Errorx(ctx)
			return global.ErrInternal
		}
		// 用户未注册
		return global.ErrUserNotRegister
	}

	// 验证密码是否正确
	if passSalt.Pass != concatPassSalt(pass, passSalt.Salt) {
		return global.ErrPassNotMatch
	}

	return nil
}
