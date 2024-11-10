package biz

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"math/rand"
	"sync"

	"github.com/google/uuid"
	"github.com/ryanreadbooks/whimer/misc/utils"
)

// 工具函数
var (
	md5ers = sync.Pool{
		New: func() any {
			return md5.New()
		},
	}
	sha256ers = sync.Pool{
		New: func() any {
			return sha256.New()
		},
	}
)

func MakeSmsCode() string {
	return fmt.Sprintf("%06d", rand.Int31n(900000)+100000)
}

func MakeRandomSalt() (s string, err error) {
	istn, err := uuid.NewUUID()
	if err != nil {
		return
	}

	suffix := utils.RandomString(8)
	prefix := utils.RandomString(8)
	candy := prefix + istn.String() + suffix

	md5er := md5ers.Get().(hash.Hash)
	defer func() {
		md5er.Reset()
		md5ers.Put(md5er)
	}()

	md5er.Write([]byte(candy))
	md5salt := md5er.Sum(nil)

	sha256er := sha256ers.Get().(hash.Hash)
	defer func() {
		sha256er.Reset()
		sha256ers.Put(sha256er)
	}()

	sha256er.Write(md5salt)
	salt := sha256er.Sum(nil)

	return hex.EncodeToString(salt), nil
}

func ConfusePassAndSalt(pass, salt string) string {
	idx := len(salt) / 2
	concat := salt[0:idx] + pass + salt[idx:]
	sha := sha256ers.Get().(hash.Hash)
	defer func() {
		sha.Reset()
		sha256ers.Put(sha)
	}()

	sha.Write([]byte(concat))
	return hex.EncodeToString(sha.Sum(nil))
}

// 生成随机初始密码
func MakeRandomInitPass() (pass string, salt string, err error) {
	salt, err = MakeRandomSalt()
	if err != nil {
		return
	}

	// 生成随机密码
	rawPass := utils.RandomPass(10) // 随机生成10位密码
	// 随机密码拼接盐
	pass = ConfusePassAndSalt(rawPass, salt)

	return
}
